package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/services"
	"github.com/mazanax/moneybot/utils"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

func HandleWithdraw(
	balanceManager *services.BalanceManager,
	stateManager *services.StateManager,
	withdrawalManager *services.WithdrawalManager,
) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		activeWithdrawal := withdrawalManager.GetActiveWithdrawal(update.CallbackQuery.From.ID)
		if activeWithdrawal != nil {
			msg := messages.WithdrawHasPending(int(activeWithdrawal.Amount/100), activeWithdrawal.Address, activeWithdrawal.Slug)
			kbd := getActiveWithdrawalKeyboard()
			if activeWithdrawal.Status == models.WithdrawalStatusProcessing {
				msg = messages.WithdrawHasProcessing(int(activeWithdrawal.Amount/100), activeWithdrawal.Address, activeWithdrawal.Slug)
				kbd = getActiveWithdrawalNoCancelKeyboard()
			}

			response := tg.NewEditMessageTextAndMarkup(
				update.CallbackQuery.From.ID,
				update.CallbackQuery.Message.MessageID,
				msg,
				kbd,
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		if balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID) < config.MinWithdrawAmount {
			response := tg.NewEditMessageTextAndMarkup(
				update.CallbackQuery.From.ID,
				update.CallbackQuery.Message.MessageID,
				messages.WithdrawLowBalance(balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID)),
				getCancelKeyboard(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}

			return
		}

		stateManager.SetState(update.CallbackQuery.From.ID, models.UserStateWithdrawAmount, update.CallbackQuery.Message.MessageID)
		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.Withdraw(balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID)),
			getCancelKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		sentMessage, err := bot.Send(response)
		if err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
		stateManager.SetState(update.CallbackQuery.From.ID, models.UserStateWithdrawAmount, sentMessage.MessageID)
	}
}

func HandleWithdrawAmount(
	balanceManager *services.BalanceManager,
	stateManager *services.StateManager,
	withdrawalManager *services.WithdrawalManager,
) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		// remove CANCEL KEYBOARD of previous message
		prevMessageID := stateManager.GetPreviousBotMessageID(update.Message.Chat.ID)
		if prevMessageID != 0 {
			response := tg.NewEditMessageText(
				update.Message.Chat.ID,
				prevMessageID,
				messages.Withdraw(balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
		}

		// process message
		amountString := strings.TrimSpace(update.Message.Text)
		if strings.ToLower(amountString) == "отмена" {
			response := tg.NewMessage(
				update.Message.Chat.ID,
				messages.Welcome(balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
			)
			response.ParseMode = tg.ModeHTML
			response.ReplyMarkup = getBalanceKeyboard()

			if _, err := bot.Send(response); err != nil {
				log.Println(err)
			}

			return
		}

		balance := balanceManager.GetCurrentBalance(update.Message.Chat.ID)
		amount, err := strconv.Atoi(amountString)
		if err != nil || amount < config.MinWithdrawAmount/100 || amount > utils.GetMaxWithdrawAmount(balance) {
			response := tg.NewMessage(
				update.Message.Chat.ID,
				messages.WithdrawIncorrectAmount(balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		stateManager.SetState(update.Message.Chat.ID, models.UserStateWithdrawAddress, 0)
		if err := withdrawalManager.StoreAmount(update.Message.Chat.ID, uint64(amount*100)); err != nil {
			log.Println(err)
			return
		}

		response := tg.NewMessage(
			update.Message.From.ID,
			messages.WithdrawConfirmAmount(amount, balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
		)
		response.ReplyMarkup = getCancelKeyboard()
		response.ParseMode = tg.ModeHTML

		sentMessage, err := bot.Send(response)
		if err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
		stateManager.SetState(update.Message.Chat.ID, models.UserStateWithdrawAddress, sentMessage.MessageID)
	}
}

func HandleWithdrawAddress(
	balanceManager *services.BalanceManager,
	stateManager *services.StateManager,
	withdrawalManager *services.WithdrawalManager,
) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			// remove message WITH CARD NUMBER
			time.Sleep(time.Second)
			deleteMessage := tg.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
			if _, err := bot.Request(deleteMessage); err != nil {
				log.Println(err)
			}

			wg.Done()
		}()

		// remove CANCEL KEYBOARD of previous message
		prevMessageID := stateManager.GetPreviousBotMessageID(update.Message.Chat.ID)
		if prevMessageID != 0 {
			response := tg.NewEditMessageText(
				update.Message.Chat.ID,
				prevMessageID,
				messages.Withdraw(balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
		}

		// process card number
		addressString := strings.TrimSpace(update.Message.Text)
		addressString = strings.ReplaceAll(addressString, " ", "")
		if strings.ToLower(addressString) == "отмена" {
			response := tg.NewMessage(
				update.Message.Chat.ID,
				messages.Welcome(balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
			)
			response.ParseMode = tg.ModeHTML
			response.ReplyMarkup = getBalanceKeyboard()

			if _, err := bot.Send(response); err != nil {
				log.Println(err)
			}

			return
		}

		cardNumber, err := strconv.Atoi(addressString)
		if err != nil || !utils.IsCardNumberValid(cardNumber) {
			response := tg.NewMessage(
				update.Message.From.ID,
				messages.WithdrawIncorrectCardNumber(),
			)
			response.ReplyMarkup = getCancelKeyboard()
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		if err := withdrawalManager.StoreAddress(update.Message.Chat.ID, addressString); err != nil {
			response := tg.NewMessage(
				update.Message.Chat.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		balance := balanceManager.GetCurrentBalance(update.Message.Chat.ID)
		draft := withdrawalManager.GetDraft(update.Message.Chat.ID)
		if draft == nil || int(draft.Amount) > balance || draft.Amount < config.MinWithdrawAmount || int(draft.Amount/100) > utils.GetMaxWithdrawAmount(balance) {
			response := tg.NewMessage(
				update.Message.Chat.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		stateManager.SetState(update.Message.Chat.ID, models.UserStateWithdrawConfirm, 0)
		response := tg.NewMessage(
			update.Message.From.ID,
			messages.WithdrawConfirmFinal(int(draft.Amount/100), addressString, balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
		)
		response.ReplyMarkup = getWithdrawalConfirmKeyboard()
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}

		wg.Wait()
	}
}

func HandleWithdrawConfirm(
	balanceManager *services.BalanceManager,
	stateManager *services.StateManager,
	withdrawalManager *services.WithdrawalManager,
	historyManager *services.HistoryManager,
) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		balance := balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID)
		draft := withdrawalManager.GetDraft(update.CallbackQuery.From.ID)
		if draft == nil || int(draft.Amount) > balance || draft.Amount < config.MinWithdrawAmount || int(draft.Amount/100) > utils.GetMaxWithdrawAmount(balance) {
			response := tg.NewMessage(
				update.CallbackQuery.Message.From.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		if err := withdrawalManager.ConfirmDraft(draft); err != nil {
			log.Println(err)
			response := tg.NewMessage(
				update.CallbackQuery.From.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
		}
		stateManager.SetState(update.CallbackQuery.From.ID, models.UserStateIdle, 0)
		balanceManager.Decrement(update.CallbackQuery.From.ID, draft.Amount+uint64(math.Floor(float64(draft.Amount)*config.Fee)))
		historyManager.CreateWithdrawal(draft)

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.WithdrawCreated(int(draft.Amount/100), draft.Address, draft.Slug),
			getActiveWithdrawalKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func HandleWithdrawCancel(withdrawalManager *services.WithdrawalManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		activeWithdrawal := withdrawalManager.GetActiveWithdrawal(update.CallbackQuery.From.ID)
		if activeWithdrawal == nil || activeWithdrawal.Status == models.WithdrawalStatusProcessing {
			response := tg.NewMessage(
				update.CallbackQuery.Message.From.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.WithdrawCancel(int(activeWithdrawal.Amount/100), activeWithdrawal.Address, activeWithdrawal.Slug),
			getWithdrawalCancelKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func HandleWithdrawCancelConfirm(
	balanceManager *services.BalanceManager,
	withdrawalManager *services.WithdrawalManager,
	historyManager *services.HistoryManager,
) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		activeWithdrawal := withdrawalManager.GetActiveWithdrawal(update.CallbackQuery.From.ID)
		if activeWithdrawal == nil || activeWithdrawal.Status == models.WithdrawalStatusProcessing {
			response := tg.NewMessage(
				update.CallbackQuery.Message.From.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		if err := withdrawalManager.CancelWithdraw(activeWithdrawal); err != nil {
			log.Println(err)
			response := tg.NewMessage(
				update.CallbackQuery.From.ID,
				messages.WithdrawUnexpectedError(),
			)
			response.ParseMode = tg.ModeHTML
			if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
		}
		balanceManager.Increment(update.CallbackQuery.From.ID, activeWithdrawal.Amount+uint64(math.Floor(float64(activeWithdrawal.Amount)*config.Fee)))
		historyManager.UpdateWithdrawal(activeWithdrawal, "Canceled by user")

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.WithdrawCanceled(
				balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID),
				int(activeWithdrawal.Amount/100),
				activeWithdrawal.Address,
				activeWithdrawal.Slug,
			),
			getActiveWithdrawalNoCancelKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func getWithdrawalConfirmKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Подтвердить", "withdraw_confirm"),
			tg.NewInlineKeyboardButtonData("Отменить", "cancel"),
		),
	)
}

func getWithdrawalCancelKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Отменить", "withdraw_cancel_confirm"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Вернуться в меню", "show_balance"),
		),
	)
}

func getActiveWithdrawalKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Отменить вывод", "withdraw_cancel"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Вернуться в меню", "show_balance"),
		),
	)
}

func getActiveWithdrawalNoCancelKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Вернуться в меню", "show_balance"),
		),
	)
}

func getCancelKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Отменить", "cancel"),
		),
	)
}

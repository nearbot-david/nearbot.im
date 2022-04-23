package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/services"
	"github.com/mazanax/moneybot/utils"
	"log"
	"strconv"
	"strings"
	"sync"
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
			msg := messages.WithdrawHasProcessing(int(activeWithdrawal.Amount), activeWithdrawal.Address, activeWithdrawal.Slug)
			kbd := getActiveWithdrawalNoCancelKeyboard()

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
		amount, err := strconv.ParseFloat(amountString, 64)
		if err != nil || amount < utils.GetMinWithdrawAmount() || amount > utils.GetMaxWithdrawAmount(balance) {
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
		if err := withdrawalManager.StoreAmount(update.Message.Chat.ID, uint64(amount*1e5)); err != nil {
			log.Println(err)
			return
		}

		response := tg.NewMessage(
			update.Message.From.ID,
			messages.WithdrawConfirmAmount(int(amount*1e5), balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
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

		// process wallet
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

		if !utils.IsNearWalletValid(addressString) {
			response := tg.NewMessage(
				update.Message.From.ID,
				messages.WithdrawIncorrectWithdrawalAddress(),
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
		if draft == nil || int(draft.Amount) > balance || float64(draft.Amount)/1e5 < utils.GetMinWithdrawAmount() || float64(draft.Amount)/1e5 > utils.GetMaxWithdrawAmount(balance) {
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
			messages.WithdrawConfirmFinal(int(draft.Amount), addressString, balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
		)
		response.ReplyMarkup = getWithdrawalConfirmKeyboard()
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
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
		if draft == nil || int(draft.Amount) > balance || float64(draft.Amount)/1e5 < utils.GetMinWithdrawAmount() || float64(draft.Amount)/1e5 > utils.GetMaxWithdrawAmount(balance) {
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
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go utils.Withdraw(draft.Address, draft.Amount, wg)

		stateManager.SetState(update.CallbackQuery.From.ID, models.UserStateIdle, 0)
		balanceManager.Decrement(update.CallbackQuery.From.ID, uint64(float64(draft.Amount)*(1+config.Fee)))
		historyManager.CreateWithdrawal(draft)

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.WithdrawCreated(int(draft.Amount), draft.Address, draft.Slug),
			getActiveWithdrawalKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}

		wg.Wait()
		withdrawalManager.MarkAsSuccessful(draft)
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

func getActiveWithdrawalKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
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

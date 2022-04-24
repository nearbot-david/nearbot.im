package handlers

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/services"
	"github.com/mazanax/moneybot/utils"
	"log"
	"strconv"
	"strings"
)

func HandleTransfer(balanceManager *services.BalanceManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		amountString := update.InlineQuery.Query
		amount, err := strconv.ParseFloat(amountString, 64)

		if amountString == "" || err != nil || amount <= 0 || int(amount*1e5) > balanceManager.GetCurrentBalance(update.InlineQuery.From.ID) {
			inlineResponse := utils.CachelessInlineConfig{
				InlineQueryID:     update.InlineQuery.ID,
				CacheTime:         0,
				IsPersonal:        true,
				SwitchPMText:      fmt.Sprintf("Available balance: %s NEAR", utils.DisplayAmount(balanceManager.GetCurrentBalance(update.InlineQuery.From.ID))),
				SwitchPMParameter: "empty_" + utils.RandStringBytes(16),
			}

			params, err := inlineResponse.Params()
			if err != nil {
				log.Println(err)
				return
			}

			if _, err := bot.MakeRequest(inlineResponse.Method(), params); err != nil {
				log.Println(err)
			}
			return
		}

		fullName := make([]string, 0)
		if update.InlineQuery.From.FirstName != "" {
			fullName = append(fullName, update.InlineQuery.From.FirstName)
		}
		if update.InlineQuery.From.LastName != "" {
			fullName = append(fullName, update.InlineQuery.From.LastName)
		}
		if update.InlineQuery.From.UserName != "" {
			if len(fullName) > 0 {
				fullName = append(fullName, fmt.Sprintf("(%s)", update.InlineQuery.From.UserName))
			} else {
				fullName = append(fullName, update.InlineQuery.From.UserName)
			}
		}

		responseArticle := tg.NewInlineQueryResultArticleHTML(
			update.InlineQuery.ID,
			fmt.Sprintf("Send %.5f NEAR (balance: %s NEAR)", amount, utils.DisplayAmount(balanceManager.GetCurrentBalance(update.InlineQuery.From.ID))),
			fmt.Sprintf("User %s sent <b>%.5f NEAR</b>.", strings.Join(fullName, " "), amount),
		)
		responseArticle.Description = fmt.Sprintf("The amount %.5f NEAR\n will be deducted from your balance NEAR. If you cancel this transfer you will get you money back.", amount)

		replyMarkup := pleaseWait()
		responseArticle.ReplyMarkup = &replyMarkup
		inlineResponse := utils.CachelessInlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{responseArticle},
		}

		params, err := inlineResponse.Params()
		if err != nil {
			log.Println(err)
			return
		}

		if _, err := bot.MakeRequest(inlineResponse.Method(), params); err != nil {
			log.Println(err)
		}
	}
}

func HandleTransferSent(balanceManager *services.BalanceManager, historyManager *services.HistoryManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		amountString := update.ChosenInlineResult.Query
		amount, err := strconv.ParseFloat(amountString, 64)

		if amountString == "" || err != nil || amount <= 0 || int(amount*1e5) > balanceManager.GetCurrentBalance(update.ChosenInlineResult.From.ID) {
			log.Println("INVALID AMOUNT")
			return
		}

		fullName := make([]string, 0)
		if update.ChosenInlineResult.From.FirstName != "" {
			fullName = append(fullName, update.ChosenInlineResult.From.FirstName)
		}
		if update.ChosenInlineResult.From.LastName != "" {
			fullName = append(fullName, update.ChosenInlineResult.From.LastName)
		}
		if update.ChosenInlineResult.From.UserName != "" {
			if len(fullName) > 0 {
				fullName = append(fullName, fmt.Sprintf("(%s)", update.ChosenInlineResult.From.UserName))
			} else {
				fullName = append(fullName, update.ChosenInlineResult.From.UserName)
			}
		}

		transfer := balanceManager.SendMoney(update.ChosenInlineResult.From.ID, 0, uint64(amount*1e5), update.ChosenInlineResult.InlineMessageID)
		if transfer == nil {
			log.Println("Cannot create transfer")
			return
		}
		historyManager.CreateTransfer(transfer)

		replyMarkup := transferKeyboard(transfer.Slug)
		updateMessage := tg.EditMessageTextConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:          update.ChosenInlineResult.From.ID,
				InlineMessageID: update.ChosenInlineResult.InlineMessageID,
				ReplyMarkup:     &replyMarkup,
			},
			Text:      fmt.Sprintf("User %s sent <b>%.5f NEAR</b>.\n\n<i>Transfer ID: %s</i>", strings.Join(fullName, " "), amount, transfer.Slug),
			ParseMode: tg.ModeHTML,
		}

		if _, err := bot.Request(updateMessage); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func HandleTransferApprove(balanceManager *services.BalanceManager, repository *repository.TransferRepository, slug string, historyManager *services.HistoryManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		transfer := repository.FindBySlug(slug)
		if transfer == nil {
			log.Printf("Cannot find transfer with slug = %s", slug)
			callback := tg.NewCallback(update.CallbackQuery.ID, fmt.Sprintf("Cannot find transfer %s. If you are sure it is an error, contact support.", slug))
			callback.ShowAlert = true
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}

			return
		}

		if transfer.From == update.CallbackQuery.From.ID {
			callback := tg.NewCallback(update.CallbackQuery.ID, "You cannot accept your transfer.")
			callback.ShowAlert = true
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}
			return
		}

		balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID) // create user balance entity
		balanceManager.Increment(update.CallbackQuery.From.ID, transfer.Amount)
		transfer.To = update.CallbackQuery.From.ID
		transfer.Status = models.TransferStatusAccepted
		if err := repository.Persist(transfer); err != nil {
			log.Println(err)

			callback := tg.NewCallback(update.CallbackQuery.ID, "Something went wrong. Please try again.")
			callback.ShowAlert = true
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}
			return
		}

		historyManager.UpdateTransfer(transfer, "Accepted")

		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		replyMarkup := checkBalanceKeyboard()
		response := tg.EditMessageTextConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:          update.CallbackQuery.From.ID,
				InlineMessageID: update.CallbackQuery.InlineMessageID,
				ReplyMarkup:     &replyMarkup,
			},
			Text:      messages.TransferAccepted(transfer.Amount, transfer.Slug),
			ParseMode: tg.ModeHTML,
		}

		if _, err := bot.Request(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func HandleTransferReject(balanceManager *services.BalanceManager, repository *repository.TransferRepository, slug string, historyManager *services.HistoryManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		transfer := repository.FindBySlug(slug)
		if transfer == nil {
			log.Printf("Cannot find transfer with slug = %s", slug)
			callback := tg.NewCallback(update.CallbackQuery.ID, fmt.Sprintf("Cannot find transfer %s. If you are sure it is an error, contact support.", slug))
			callback.ShowAlert = true
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}

			return
		}

		// sender canceled transfer
		if transfer.From == update.CallbackQuery.From.ID {
			balanceManager.Increment(update.CallbackQuery.From.ID, transfer.Amount)
			transfer.Status = models.TransferStatusCanceled
			if err := repository.Persist(transfer); err != nil {
				log.Println(err)

				callback := tg.NewCallback(update.CallbackQuery.ID, "Произошла какая-то ошибка. Попробуйте еще раз.")
				callback.ShowAlert = true
				if _, err := bot.Request(callback); err != nil {
					log.Println(err)
					return
				}
				return
			}

			historyManager.UpdateTransfer(transfer, "Canceled by sender")
			callback := tg.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}

			replyMarkup := checkBalanceKeyboard()
			response := tg.EditMessageTextConfig{
				BaseEdit: tg.BaseEdit{
					ChatID:          update.CallbackQuery.From.ID,
					InlineMessageID: update.CallbackQuery.InlineMessageID,
					ReplyMarkup:     &replyMarkup,
				},
				Text:      messages.TransferCanceled(transfer.Amount, transfer.Slug),
				ParseMode: tg.ModeHTML,
			}

			if _, err := bot.Request(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
				log.Println(err)
			}
			return
		}

		balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID) // create user balance entity
		balanceManager.Increment(transfer.From, transfer.Amount)
		transfer.To = update.CallbackQuery.From.ID
		transfer.Status = models.TransferStatusRejected
		if err := repository.Persist(transfer); err != nil {
			log.Println(err)

			callback := tg.NewCallback(update.CallbackQuery.ID, "Something went wrong. Please try again.")
			callback.ShowAlert = true
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}
			return
		}

		historyManager.UpdateTransfer(transfer, "Rejected by receiver")
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		replyMarkup := checkBalanceKeyboard()
		response := tg.EditMessageTextConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:          update.CallbackQuery.From.ID,
				InlineMessageID: update.CallbackQuery.InlineMessageID,
				ReplyMarkup:     &replyMarkup,
			},
			Text:      messages.TransferRejected(transfer.Amount, transfer.Slug),
			ParseMode: tg.ModeHTML,
		}

		if _, err := bot.Request(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func transferKeyboard(transferID string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Accept", "transfer_approve_"+transferID),
			tg.NewInlineKeyboardButtonData("Decline", "transfer_reject_"+transferID),
		),
	)
}

func pleaseWait() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Processing transfer...", "no_data"),
		),
	)
}

func checkBalanceKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonURL("Show balance", "https://t.me/textmoneybot"),
		),
	)
}

package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/services"
	"log"
	"strings"
)

func HandleBalance(balanceManager *services.BalanceManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.Balance(balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID)),
			getBalanceKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func HandleAddress(balanceManager *services.BalanceManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		address, balance := balanceManager.GetAddressBalance(update.CallbackQuery.From.ID)
		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.Address(address, balance),
			getBalanceKeyboard(),
		)
		response.DisableWebPagePreview = false
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func getBalanceKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Balance", "show_balance"),
			tg.NewInlineKeyboardButtonData("Address", "show_address"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Top Up", "deposit"),
			tg.NewInlineKeyboardButtonData("Transfer", "withdraw"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("History", "history"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Support", "support"),
		),
	)
}

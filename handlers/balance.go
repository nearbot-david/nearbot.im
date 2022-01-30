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

func getBalanceKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Баланс", "show_balance"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Пополнить", "deposit"),
			tg.NewInlineKeyboardButtonData("Вывести", "withdraw"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("История", "history"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Поддержка", "support"),
		),
	)
}

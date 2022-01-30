package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/messages"
	"log"
	"strings"
)

func HandleUnsupportedMessage() HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		var response tg.EditMessageTextConfig

		if update.CallbackData() != "" {
			callback := tg.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}

			response = tg.NewEditMessageTextAndMarkup(
				update.CallbackQuery.From.ID,
				update.CallbackQuery.Message.MessageID,
				messages.UnsupportedMessage(),
				getMenuKeyboard(),
			)
		} else {
			response = tg.NewEditMessageTextAndMarkup(
				update.Message.From.ID,
				update.Message.MessageID,
				messages.UnsupportedMessage(),
				getMenuKeyboard(),
			)
		}

		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func getMenuKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Вернуться в меню", "show_balance"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Поддержка", "support"),
		),
	)
}

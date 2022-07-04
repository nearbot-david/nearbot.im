package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/messages"
	"log"
	"strings"
)

func HandleSupport() HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.Support(),
			getSupportKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func getSupportKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonURL("Frequently asked questions", "https://nearbot.im/faq"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonURL("Contact us", "https://t.me/textmoney_support"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Return to menu", "cancel"),
		),
	)
}

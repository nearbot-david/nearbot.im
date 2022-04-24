package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleHistory() HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "History not available right now, but we're working on it!")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}
	}
}

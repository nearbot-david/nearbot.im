package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleHistory() HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "История пока недоступна, но мы уже работаем над этим")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}
	}
}

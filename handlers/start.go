package handlers

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/services"
	"log"
)

func HandleStart(balanceManager *services.BalanceManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		response := tg.NewMessage(
			update.Message.Chat.ID,
			messages.Welcome(balanceManager.GetCurrentBalance(update.Message.Chat.ID)),
		)
		response.ParseMode = tg.ModeHTML
		response.ReplyMarkup = getBalanceKeyboard()

		if _, err := bot.Send(response); err != nil {
			log.Println(err)
		}
	}
}

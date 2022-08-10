package handlers

import (
	"github.com/Pay-With-NEAR/nearbot.im/messages"
	"github.com/Pay-With-NEAR/nearbot.im/models"
	"github.com/Pay-With-NEAR/nearbot.im/services"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleCancel(balanceManager *services.BalanceManager, stateManager *services.StateManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "Action canceled")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		stateManager.SetState(update.CallbackQuery.From.ID, models.UserStateIdle, update.CallbackQuery.Message.MessageID)
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

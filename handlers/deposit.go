package handlers

import (
	"fmt"
	"github.com/Pay-With-NEAR/nearbot.im/messages"
	"github.com/Pay-With-NEAR/nearbot.im/repository"
	"github.com/Pay-With-NEAR/nearbot.im/services"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleDeposit(balanceManager *services.BalanceManager) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		callback := tg.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Println(err)
			return
		}

		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.Deposit(balanceManager.GetCurrentBalance(update.CallbackQuery.From.ID)),
			getDepositKeyboard(),
		)
		response.ParseMode = tg.ModeHTML
		if _, err := bot.Send(response); err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}
	}
}

func HandleDepositAmount(paymentMethod services.PaymentMethod, depositRepository *repository.DepositRepository, amount uint64) HandlerFunc {
	return func(bot *tg.BotAPI, update *tg.Update) {
		if err := validateAmount(amount); err != nil {
			callback := tg.NewCallbackWithAlert(update.CallbackQuery.ID, "Unknown amount!")
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}
		} else {
			callback := tg.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				log.Println(err)
				return
			}
		}

		paymentID, paymentLink := paymentMethod.GeneratePaymentLink(update.CallbackQuery.From.ID, amount)
		response := tg.NewEditMessageTextAndMarkup(
			update.CallbackQuery.From.ID,
			update.CallbackQuery.Message.MessageID,
			messages.DepositAmount(float64(amount)/100*1e5, string(paymentLink), string(paymentID)),
			getDepositAmountKeyboard(string(paymentLink)),
		)
		response.ParseMode = tg.ModeHTML
		sentMessage, err := bot.Send(response)
		if err != nil && !strings.Contains(err.Error(), "message content and reply markup are exactly the same") {
			log.Println(err)
		}

		deposit := depositRepository.FindBySlug(string(paymentID))
		if deposit != nil {
			deposit.MessageID = sentMessage.MessageID
			_ = depositRepository.Persist(deposit)
		}
	}
}

func validateAmount(amount uint64) error {
	for _, a := range []uint64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000} {
		if a == amount {
			return nil
		}
	}

	return fmt.Errorf("invalid amount")
}

func getDepositKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("0.1 NEAR", "deposit_10"),
			tg.NewInlineKeyboardButtonData("0.25 NEAR", "deposit_25"),
			tg.NewInlineKeyboardButtonData("0.5 NEAR", "deposit_50"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("1 NEAR", "deposit_100"),
			tg.NewInlineKeyboardButtonData("2.5 NEAR", "deposit_250"),
			tg.NewInlineKeyboardButtonData("5 NEAR", "deposit_500"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("10 NEAR", "deposit_1000"),
			tg.NewInlineKeyboardButtonData("25 NEAR", "deposit_2500"),
			tg.NewInlineKeyboardButtonData("50 NEAR", "deposit_5000"),
		),
		//TODO: implement me later
		//tg.NewInlineKeyboardRow(
		//	tg.NewInlineKeyboardButtonData("???????????? ??????????", "deposit_other"),
		//),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Cancel", "cancel"),
		),
	)
}

func getDepositAmountKeyboard(paymentLink string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonURL("Pay with NEAR", paymentLink)),
		tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData("Cancel", "cancel")),
	)
}

package handlers

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/services"
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
			callback := tg.NewCallbackWithAlert(update.CallbackQuery.ID, "Неизвестная сумма!")
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
			messages.DepositAmount(amount, string(paymentLink), string(paymentID)),
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
	for _, a := range []uint64{100, 500, 1000, 2500, 5000, 10000} {
		if a == amount {
			return nil
		}
	}

	return fmt.Errorf("invalid amount")
}

func getDepositKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("100₽", "deposit_100"),
			tg.NewInlineKeyboardButtonData("500₽", "deposit_500"),
			tg.NewInlineKeyboardButtonData("1000₽", "deposit_1000"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("2500₽", "deposit_2500"),
			tg.NewInlineKeyboardButtonData("5000₽", "deposit_5000"),
			tg.NewInlineKeyboardButtonData("10 000₽", "deposit_10000"),
		),
		//TODO: implement me later
		//tg.NewInlineKeyboardRow(
		//	tg.NewInlineKeyboardButtonData("Другая сумма", "deposit_other"),
		//),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Отменить", "cancel"),
		),
	)
}

func getDepositAmountKeyboard(paymentLink string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonURL("Перейти к оплате", paymentLink)),
		tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData("Отменить", "cancel")),
	)
}

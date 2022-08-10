package endpoints

import (
	"fmt"
	"github.com/Pay-With-NEAR/nearbot.im/config"
	"github.com/Pay-With-NEAR/nearbot.im/messages"
	"github.com/Pay-With-NEAR/nearbot.im/services"
	"github.com/Pay-With-NEAR/nearbot.im/utils"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"sync"
	"time"
)

func PaymentEndpoint(
	paymentMethod services.PaymentMethod,
	balanceManager *services.BalanceManager,
	historyManager *services.HistoryManager,
	addressManager *services.AddressManager,
	bot *tg.BotAPI,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			handlePost(paymentMethod, balanceManager, historyManager, addressManager, bot)(writer, request)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
			writer.Write(errorPage())
		}
	}
}

func PaymentSuccessfulEndpoint() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write(paymentSuccessfulPage())
	}
}

func PaymentFailedEndpoint() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write(paymentFailedPage())
	}
}

func handlePost(
	paymentMethod services.PaymentMethod,
	balanceManager *services.BalanceManager,
	historyManager *services.HistoryManager,
	addressManager *services.AddressManager,
	bot *tg.BotAPI,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		deposit, err := paymentMethod.ProcessPayment(request)
		if err == nil {
			addr, _ := balanceManager.GetAddressBalance(deposit.TelegramID) // just to be sure that user already has address

			balanceManager.Increment(deposit.TelegramID, deposit.Amount)
			historyManager.CreateDeposit(deposit)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go addressManager.Transfer(config.NearWallet, addr, deposit.Amount, wg)

			if deposit.MessageID != 0 {
				paymentLink := paymentMethod.BuildPaymentLink(services.PaymentID(deposit.Slug))

				// скрываем клавиатуру, но оставляем текст сообщения старый
				text := messages.DepositAmount(float64(deposit.Amount), string(paymentLink), deposit.Slug)
				hideMarkup := tg.NewEditMessageText(deposit.TelegramID, deposit.MessageID, text)
				hideMarkup.ParseMode = tg.ModeHTML
				_, _ = bot.Send(hideMarkup)
			}

			message := tg.NewMessage(deposit.TelegramID, messages.DepositProcessed(balanceManager.GetCurrentBalance(deposit.TelegramID), deposit.Slug))
			message.ParseMode = tg.ModeHTML
			bot.Send(message)
			writer.Write([]byte(""))

			wg.Add(1)
			go func() {
				time.Sleep(time.Second / 2)

				message := tg.NewMessage(
					deposit.TelegramID,
					messages.Balance(balanceManager.GetCurrentBalance(deposit.TelegramID)),
				)
				message.ParseMode = tg.ModeHTML
				message.ReplyMarkup = getBalanceKeyboard()
				bot.Send(message)

				wg.Done()
			}()

			wg.Wait()
			return
		}

		fmt.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
	}
}

func errorPage() []byte {
	return utils.RenderTemplate(errorTemplate, "")
}

func paymentSuccessfulPage() []byte {
	return utils.RenderTemplate(paymentSuccessfulTemplate, "")
}

func paymentFailedPage() []byte {
	return utils.RenderTemplate(paymentFailedTemplate, "")
}

func getBalanceKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Balance", "show_balance"),
			tg.NewInlineKeyboardButtonData("Address", "show_address"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Top up", "deposit"),
			tg.NewInlineKeyboardButtonData("Transfer", "withdraw"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("History", "history"),
		),
	)
}

const errorTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Something wrong &bull; @sendnearbot</title>
    <style>
		body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
	</style>
</head>
<body>
<div class="wrapper">
    <header><h1>Error &bull; @sendnearbot</h1></header>

    <section>
        <p>There was an error while making this request. Please make sure you followed the correct link.</p>
		<p>If the error repeats, we recommend contacting support.</p>

		<a class="block" href="https://t.me/sendnearbot">Return to bot</a>
    </section>
</div>
</body>
</html>`

const paymentSuccessfulTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Successful payment &bull; @sendnearbot</title>
    <style>
        body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
    </style>
</head>
<body>
<div class="wrapper">
    <header><h1>Successful payment &bull; @sendnearbot</h1></header>
    <section>
        <p>Successful payment.</p>

        <p>You will get message from bot when NEAR will be added to your balance.</p>

        <p>Usually it takes 1-2 minutes.</p>

        <a class="block" href="https://t.me/sendnearbot">Return to bot</a>
    </section>
</div>
</body>
</html>`

const paymentFailedTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Payment failed &bull; @sendnearbot</title>
    <style>
        body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
    </style>
</head>
<body>
<div class="wrapper">
    <header><h1>Payment failed &bull; @sendnearbot</h1></header>
    <section>
        <p>Payment failed.</p>

        <p>Please return to bot and try again.</p>

        <a class="block" href="https://t.me/sendnearbot">Return to bot</a>
    </section>
</div>
</body>
</html>`

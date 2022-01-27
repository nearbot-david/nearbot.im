package endpoints

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/messages"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/services"
	"github.com/mazanax/moneybot/utils"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func PaymentEndpoint(
	paymentMethod *services.GatewayPaymentMethod,
	depositRepository *repository.DepositRepository,
	balanceManager *services.BalanceManager,
	bot *tg.BotAPI,
	isDebug bool,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			handlePost(paymentMethod, balanceManager, bot)(writer, request)
		case http.MethodGet:
			handleGet(paymentMethod, depositRepository, isDebug)(writer, request)
		default:
			writer.WriteHeader(http.StatusMethodNotAllowed)
			writer.Write(errorPage())
		}
	}
}

func handlePost(
	paymentMethod services.PaymentMethod,
	balanceManager *services.BalanceManager,
	bot *tg.BotAPI,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		deposit, err := paymentMethod.ProcessPayment(request)
		if err == nil {
			balanceManager.Increment(deposit.TelegramID, deposit.Amount)

			if deposit.MessageID != 0 {
				paymentLink := paymentMethod.BuildPaymentLink(services.PaymentID(deposit.Slug))

				text := messages.DepositAmount(deposit.Amount/100, string(paymentLink), deposit.Slug)
				hideMarkup := tg.NewEditMessageText(deposit.TelegramID, deposit.MessageID, text)
				hideMarkup.ParseMode = tg.ModeHTML
				_, _ = bot.Send(hideMarkup)
			}

			message := tg.NewMessage(deposit.TelegramID, messages.DepositProcessed(balanceManager.GetCurrentBalance(deposit.TelegramID), deposit.Slug))
			message.ParseMode = tg.ModeHTML
			bot.Send(message)
			writer.Write([]byte(""))

			wg := &sync.WaitGroup{}
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

		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
	}
}

func handleGet(
	paymentMethod *services.GatewayPaymentMethod,
	depositRepository *repository.DepositRepository,
	isDebug bool,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		txID := request.URL.Path[len("/payment/"):]
		if txID == "" {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(errorPage())
			return
		}

		deposit := depositRepository.FindBySlug(txID)
		if deposit == nil {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write(errorPage())
			return
		}
		if deposit.Status != models.PaymentStatusNew {
			writer.Write(paymentSuccessfulPage(deposit.Slug, uint(deposit.Amount), deposit.UpdatedAt))
			return
		}

		amount := deposit.Amount + uint64(math.Floor(float64(deposit.Amount)*config.Fee))
		form := paymentMethod.GetPaymentForm(txID, amount)

		writer.WriteHeader(200)
		writer.Write(paymentPage(txID, uint(deposit.Amount), form, isDebug))
	}
}

func paymentPage(txID string, amount uint, form string, isDebug bool) []byte {
	data := struct {
		TxID          string
		Amount        string
		Fee           string
		AmountWithFee string
		Form          template.HTML
		Debug         bool
	}{
		TxID:          txID,
		Amount:        strconv.Itoa(int(amount / 100)),
		Fee:           strconv.Itoa(int(math.Floor(float64(amount)*config.Fee) / 100)),
		AmountWithFee: strconv.Itoa(int(amount/100) + int(math.Floor(float64(amount)*config.Fee)/100)),
		Form:          template.HTML(form),
		Debug:         isDebug,
	}

	return utils.RenderTemplate(paymentTemplate, data)
}

func paymentSuccessfulPage(txID string, amount uint, updatedAt time.Time) []byte {
	data := struct {
		TxID          string
		Amount        string
		Fee           string
		AmountWithFee string
		UpdatedAt     string
	}{
		TxID:          txID,
		Amount:        strconv.Itoa(int(amount / 100)),
		Fee:           strconv.Itoa(int(math.Floor(float64(amount)*config.Fee) / 100)),
		AmountWithFee: strconv.Itoa(int(amount/100) + int(math.Floor(float64(amount)*config.Fee)/100)),
		UpdatedAt:     updatedAt.Format("2006-01-02 3:04PM"),
	}

	return utils.RenderTemplate(paymentSuccessfulTemplate, data)
}

func errorPage() []byte {
	return utils.RenderTemplate(errorTemplate, "")
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
	)
}

const paymentTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Пополнение баланса &bull; @textmoneybot</title>
    <style>
		body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
	</style>
</head>
<body>
<div class="wrapper">
    <header><h1>Пополнение баланса &bull; @textmoneybot</h1></header>

    <section>
        <p>Вам выставлен счет на оплату.</p>
        <p>
            <b>Номер счета:</b> {{ .TxID }}<br>
			<b>Сумма:</b> {{ .Amount }}&#8381;<br>
			<b>Комиссия:</b> {{ .Fee }}&#8381;<br>
			<b>Сумма с учетом комиссии:</b> {{ .AmountWithFee }}&#8381;<br><br>

            <b>Способ оплаты:</b> Банковская карта
        </p>

		{{ if .Debug }}
			<a class="block" href="/emulator/{{ .TxID }}" title="Используется эмулятор">Оплатить (E)</a>

			<p style="text-align: center;"><small>Бот запущен в демо-режиме. Оплата будет проведена через эмулятор.</small></p>
		{{ else }}
        	{{ .Form }}
		{{ end }}
    </section>
</div>
</body>
</html>`

const paymentSuccessfulTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Баланс пополнен &bull; @textmoneybot</title>
    <style>
		body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
	</style>
</head>
<body>
<div class="wrapper">
    <header><h1>Баланс пополнен &bull; @textmoneybot</h1></header>

    <section>
        <p>Счет успешно оплачен. Деньги зачислены на ваш баланс.</p>
        <p>
            <b>Номер счета:</b> {{ .TxID }}<br>
			<b>Сумма:</b> {{ .Amount }}&#8381;<br>
			<b>Комиссия:</b> {{ .Fee }}&#8381;<br>
			<b>Сумма с учетом комиссии:</b> {{ .AmountWithFee }}&#8381;<br><br>

            <b>Способ оплаты:</b> Банковская карта<br>
			<b>Дата и время:</b> {{ .UpdatedAt }}
        </p>

		<a class="block" href="https://t.me/textmoneybot">Вернуться в бота</a>
    </section>
</div>
</body>
</html>`

const errorTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Ошибка &bull; @textmoneybot</title>
    <style>
		body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
	</style>
</head>
<body>
<div class="wrapper">
    <header><h1>Ошибка &bull; @textmoneybot</h1></header>

    <section>
        <p>При выполнении запроса произошла ошибка. Пожалуйста, убедитесь, что вы перешли по правильной ссылке.</p>
		<p>Если ошибка повторяется, рекомендуем обратиться в поддержку.</p>

		<a class="block" href="https://t.me/textmoneybot">Вернуться в бота</a>
    </section>
</div>
</body>
</html>`

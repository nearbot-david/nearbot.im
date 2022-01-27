package endpoints

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/utils"
	"math"
	"net/http"
	"strconv"
)

func EmulatorEndpoint(depositRepository *repository.DepositRepository, secretKey string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		txID := request.URL.Path[len("/emulator/"):]
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

		writer.WriteHeader(200)
		writer.Write(emulatorPage(
			txID,
			deposit.Amount,
			fmt.Sprintf("%d", deposit.TelegramID),
			secretKey,
			string(deposit.Status),
		))
	}
}

func emulatorPage(txID string, amount uint64, userID string, secretKey string, status string) []byte {
	signature := md5.Sum([]byte(fmt.Sprintf("UUID:%s,Amount:%d,SecretKey:%s", txID, int(amount)+int(math.Floor(float64(amount)*config.Fee)), secretKey)))
	data := struct {
		TxID          string
		UserID        string
		Amount        string
		Fee           string
		AmountWithFee string
		Signature     string
		Status        string
	}{
		TxID:          txID,
		UserID:        userID,
		Amount:        fmt.Sprintf("%.2f", float64(amount)/100),
		Fee:           strconv.Itoa(int(math.Floor(float64(amount)*config.Fee) / 100)),
		AmountWithFee: strconv.Itoa(int(amount/100) + int(math.Floor(float64(amount)*config.Fee)/100)),
		Signature:     hex.EncodeToString(signature[:]),
		Status:        status,
	}

	return utils.RenderTemplate(emulatorTemplate, data)
}

const emulatorTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Emulator &bull; @textmoneybot</title>
    <style>
        body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}
    </style>
</head>
<body>
<div class="wrapper">
    <header><h1>Платеж {{ .TxID }}</h1></header>

    <section>
		<p style="color: red;" id="message"></p>

        <p>
            <b>ID пользователя:</b> {{ .UserID }}<br>
            <b>Сумма:</b> {{ .Amount }}&#8381;<br>
            <b>Комиссия:</b> {{ .Fee }}&#8381;<br>
            <b>Сумма с учетом комиссии:</b> {{ .AmountWithFee }}&#8381;<br><br>
            <b>Статус:</b> {{ .Status }}
        </p>

		{{ if ne .Status "SUCCESS" }}
        <a id="process-payment" class="block" href="javascript://">Оплатить</a>
		{{ end }}
    </section>
</div>
<script>
    document.querySelector('#process-payment').onclick = function () {
		const formData = new FormData();
		formData.append('uuid', '{{ .TxID }}');
		formData.append('amount', Math.round({{ .AmountWithFee }} * 100));
		formData.append('signature', '{{ .Signature }}');

        fetch('/payment/', {
            method: 'POST',
            body: formData
        })
        .then(response => {
			if (response.status !== 200) {
				document.getElementById('message').innerHTML = 'Ошибка обработки';
			}
			setTimeout(() => location.reload(), 1000);
        })
    }
</script>
</body>
</html>`

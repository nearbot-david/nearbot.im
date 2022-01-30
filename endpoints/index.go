package endpoints

import (
	"github.com/mazanax/moneybot/utils"
	"net/http"
)

func IndexEndpoint() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		writer.WriteHeader(200)
		writer.Write(indexPage())
	}
}

func indexPage() []byte {
	return utils.RenderTemplate(indexTemplate, "")
}

const indexTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>@textmoneybot</title>
    <style>
        body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}.video{width:75%;max-width:450px;background:#fff;'}
    </style>
</head>
<body>
<div class="wrapper">
    <header><h1>Textmoneybot</h1></header>

    <section style="text-align: center;">
		<p style="font-style: italic;">Удобный способ перевести деньги любому пользователю Telegram, не покидая мессенджер.</p>

		<p>
			<video loop muted autoplay class="video">
				<source src="https://textmoney.mznx.dev/demo.mp4?v4" type="video/mp4"/>
            </video>
		</p>

		<p>Пополняйте баланс с банковской карты, переводите деньги друзьям и знакомым, не выходя из Telegram.</p>

		<a class="block" href="https://t.me/textmoneybot">Перейти в бота</a>

		<p><a href="/faq">Часто задаваемые вопросы</a></p>
    </section>
</div>
</body>
</html>`

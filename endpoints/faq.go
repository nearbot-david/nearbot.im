package endpoints

import (
	"github.com/mazanax/moneybot/utils"
	"net/http"
)

func FaqEndpoint() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		writer.WriteHeader(200)
		writer.Write(faqPage())
	}
}

func faqPage() []byte {
	return utils.RenderTemplate(faqTemplate, "")
}

const faqTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>FAQ &bull; @textmoneybot</title>
    <style>
        body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}.video{width:75%;max-width:450px;background:#fff;}
    </style>
</head>
<body>
<div class="wrapper">
    <header><h1>FAQ</h1></header>

        <section>
        <p><a href="/">&larr; Вернуться на главную</a></p>

        <ul>
            <li><a href="#what-is-it">Для чего этот бот?</a></li>
            <li><a href="#how-it-works">Как это работает?</a></li>
            <li><a href="#fees">Какие есть комиссии?</a></li>
            <li><a href="#withdraws">Как вывести деньги?</a></li>
            <li><a href="#accept-by-myself">Могу ли я принять свой же перевод?</a></li>
            <li><a href="#cancel-by-myself">Могу ли я отменить свой перевод?</a></li>
            <li><a href="#contact-support">Контакты для связи</a></li>
        </ul>

        <h4 id="what-is-it">Для чего этот бот?</h4>
        <p>С помощью Textmoneybot вы можете переводить деньги своим друзьям и знакомым через Telegram.</p>
        <hr>

        <h4 id="how-it-works">Как это работает?</h4>
        <p>Все просто: вы пополняете баланс с помощью банковской карты, после чего можете перевести деньги любому
            пользователю Telegram.</p>
        <p>Для этого в чате с пользователем напишите @textmoneybot и сумму, которую хотите
            перевести,
            если на вашем балансе достаточно денег, вы увидете кнопку, по нажатию на которую будет отправлено сообщение
            с переводом.</p>
        <p style="text-align: center;">
            <video loop muted autoplay class="video">
                <source src="https://textmoney.mznx.dev/demo.mp4?v4" type="video/mp4"/>
            </video>
        </p>
        <hr>

        <h4 id="fees">Какие есть комиссии?</h4>
        <p>В данный момент комиссия за пополнение и вывод составляет 5%. Комиссии за перевод отсутствуют.</p>
        <hr>

        <h4 id="withdraws">Как вывести деньги?</h4>
        <p>Чтобы вывести деньги, нажмите в меню кнопку <b>Вывести</b> и следуйте инструкциям. Обычно время обработки
            запроса на вывод составляет 1-2 рабочих дня.</p>
        <p><b>Обратите внимание:</b> одновременно может быть только одна заявка на вывод, но вы можете отменить
            существующую (если она еще ожидает обработки) и создать новую.</p>
        <hr>

        <h4 id="accept-by-myself">Могу ли я принять свой же перевод?</h4>
        <p>Нет. Пользователь, создавший перевод не может его же принять.</p>
        <hr>

        <h4 id="cancel-by-myself">Могу ли я отменить свой перевод?</h4>
        <p>Да. Для этого нажмите на кнопку отклонить. Деньги вернутся на ваш баланс.</p>
        <p style="text-decoration: line-through;" title="В разработке">Кроме того, вы можете отменить перевод из истории операций. Для этого нажмите на соответствующую кнопку рядом
            с переводом.</p>
        <hr>

        <h4 id="contact-support">У меня еще остались вопросы</h4>
        <p>Если у вас еще остались вопросы, напишите нам в телеграм <a href="https://t.me/textmoney_support">@textmoney_support</a>.</p>
        <hr>

        <p><a href="/">&larr; Вернуться на главную</a></p>
    </section>
</div>
</body>
</html>`

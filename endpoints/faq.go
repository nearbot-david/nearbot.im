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
		_, err := writer.Write(faqPage())
		if err != nil {
			return
		}
	}
}

func faqPage() []byte {
	return utils.RenderTemplate(faqTemplate, "")
}

const faqTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>FAQ &bull; @sendnearbot</title>
    <style>
        body,html{margin:0;padding:0;font-size:18px}.wrapper{width:100%;max-width:800px;margin:0 auto}header{background:#f0f8ff;padding:15px 30px;text-align:center;color:#000;text-shadow:1px 1px 2px #fff;margin:0}header h1{font-size:24px}section{padding:5px}section a{color:#888;text-decoration:none}section a:hover{border-bottom:1px dashed #888}section.logo{margin:5px auto;max-width:240px;width:100%}section.logo img{width:100%}section .block,section .mznx-payment-button{border:none!important;font-size:18px;margin:15px auto 0;padding:15px 30px;width:50%;background:#88bbf8;display:block;text-align:center;color:#fff;text-shadow:1px 1px 2px #888}section .block:hover,section .mznx-payment-button:hover{background:#5e84ad}.video{width:75%;max-width:450px;background:#fff;}
    </style>
</head>
<body>
<div class="wrapper">
    <header><h1>FAQ</h1></header>

        <section>
        <p><a href="/">&larr; Return to homepage</a></p>

        <ul>
            <li><a href="#what-is-it">What is it?</a></li>
            <li><a href="#how-it-works">How it works?</a></li>
            <li><a href="#fees">Do you have fees?</a></li>
            <li><a href="#withdraws">How to withdraw NEAR?</a></li>
            <li><a href="#accept-by-myself">Can I accept my transfer to another user?</a></li>
            <li><a href="#cancel-by-myself">Can I cancel my transfer to another user?</a></li>
            <li><a href="#contact-support">Contact</a></li>
        </ul>

        <h4 id="what-is-it">What is it?</h4>
        <p>This bot makes it possible to send NEAR to any Telegram user even if they don't have NEAR wallet.</p>
        <hr>

        <h4 id="how-it-works">How it works?</h4>
        <p>It's easy: you top up bot balance from you NEAR wallet and then you can send your NEAR to anyone.</p>
        <p>To do it you need to write @sendnearbot and amount of NEAR you're going to send in chat with user you want to send NEAR to.
            If you have enough NEAR on your balance you will see button.</p>
        <p style="text-align: center;">
            <video loop muted autoplay class="video">
                <source src="https://nearbot.im/demo-near.mp4?v3" type="video/mp4"/>
            </video>
        </p>
        <hr>

        <h4 id="fees">Do you have fees?</h4>
        <p>At the moment we have only withdrawal fee 5%.</p>
        <hr>

        <h4 id="withdraws">How to withdraw NEAR?</h4>
        <p>Click button <b>withdraw</b> in the bot menu and follow instructions.</p>
        <hr>

        <h4 id="accept-by-myself">Can I accept my transfer to another user?</h4>
        <p>No, you can't</p>
        <hr>

        <h4 id="cancel-by-myself">Can I cancel my transfer to another user?</h4>
        <p>Yes. Just click <b>decline</b> under message with transfer and you will get your NEAR back</p>
        <hr>

        <h4 id="contact-support">I still have questions...</h4>
        <p>If you still have questions, feel free to write us: <a href="https://t.me/textmoney_support">@textmoney_support</a>.</p>
        <hr>

        <p><a href="/">&larr; Return to homepage</a></p>
    </section>
</div>
</body>
</html>`

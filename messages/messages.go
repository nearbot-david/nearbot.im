package messages

import (
	"fmt"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/utils"
	"time"
)

func Welcome(balance int) string {
	return "<b>Добро пожаловать!</b>\n\n" +
		"С помощью этого бота вы можете перевести NEAR любому пользователю Telegram," +
		" просто отправив сообщение с суммой.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>", utils.DisplayAmount(balance))
}

func Balance(balance int) string {
	return "<b>TextMoney • NEAR</b>\n\n" +
		"С помощью этого бота вы можете перевести NEAR любому пользователю Telegram," +
		" просто отправив сообщение с суммой.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Обновлено: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func Deposit(balance int) string {
	return "<b>Пополнение баланса.</b>\n\n" +
		"Выберите сумму, которую хотите внести на баланс.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Обновлено: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func DepositAmount(amount float64, paymentLink string, paymentID string) string {
	return "<b>Пополнение баланса.</b>\n\n" +
		fmt.Sprintf("Ваш баланс будет пополнен на <b>%s NEAR</b>. Продолжить?\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("Ссылка для оплаты: %s\n\n", paymentLink) +
		fmt.Sprintf("<i>ID платежа: %s</i>", paymentID)
}

func DepositProcessed(balance int, paymentID string) string {
	return "<b>Платеж успешно зачислен.</b>\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Обновлено: %s</i>\n\n", time.Now().Format("2006-01-02 3:04PM")) +
		fmt.Sprintf("<i>ID платежа: %s</i>", paymentID)
}

func Withdraw(balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		fmt.Sprintf("Укажите сумму вывода. Обратите внимание, что комиссия составит %.1f%% от суммы вывода.\n\n", config.Fee*100) +
		fmt.Sprintf("<b>Минимальная сумма:</b> %.5f NEAR\n", utils.GetMinWithdrawAmount()) +
		fmt.Sprintf("<b>Максимальная сумма:</b> %.5f NEAR\n\n", calculateMaxWithdrawAmount(balance)) +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawLowBalance(balance int) string {
	return "<b>Вывод недоступен</b>\n\n" +
		"На вашем балансе недостаточно NEAR для вывода.\n" +
		fmt.Sprintf("<b>Минимальная сумма:</b> %.5f NEAR\n\n", utils.GetMinWithdrawAmount()) +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawIncorrectAmount(balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		"Укажите сумму вывода в NEAR. Например 0.2\nДля отмены отправьте текст \"отмена\"\n\n" +
		fmt.Sprintf("<b>Минимальная сумма:</b> %.5f NEAR\n", utils.GetMinWithdrawAmount()) +
		fmt.Sprintf("<b>Максимальная сумма:</b> %.5f NEAR\n\n", calculateMaxWithdrawAmount(balance)) +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawConfirmAmount(amount int, balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		fmt.Sprintf("Вы собираетесь вывести <b>%s NEAR</b>. Пришлите адрес кошелька, на который хотите вывести NEAR.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Например: textmoneybot.near\n\n") +
		fmt.Sprintf("<b>Комиссия:</b> %s NEAR\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("С вашего баланса будет списано <b>%s NEAR</b> (с учетом комиссии)\n\n", utils.DisplayAmount(int(float64(amount)*(1+config.Fee)))) +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawConfirmFinal(amount int, wallet string, balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		fmt.Sprintf("Вы собираетесь вывести <b>%s NEAR</b>.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Адрес <b>%s</b>.\n\n", wallet) +
		fmt.Sprintf("<b>Комиссия:</b> %s NEAR\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("Продолжить? С вашего баланса будет списано <b>%s NEAR</b> (с учетом комиссии)\n\n", utils.DisplayAmount(int(float64(amount)*(1+config.Fee)))) +
		fmt.Sprintf("Текущий баланс: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawHasProcessing(amount int, wallet string, withdrawalID string) string {
	return "<b>Активный вывод</b>\n\n" +
		fmt.Sprintf("Прямо сейчас у вас есть один запрос на вывод, находящийся в обработке. К сожалению, отменить вывод находящийся в обработке невозможно.\n\n") +
		fmt.Sprintf("Статус: <b>ожидает обработки</b>\n") +
		fmt.Sprintf("Сумма вывода: <b>%s NEAR</b>.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Адрес <b>%s</b>.\n\n", wallet) +
		fmt.Sprintf("Комиссия: <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func WithdrawCreated(amount int, wallet string, withdrawalID string) string {
	return "<b>Запрос на вывод создан</b>\n\n" +
		fmt.Sprintf("Обратите внимание, вы можете создать только один запрос на вывод за раз. Чтобы создать новый, дождитесь окончания обработки. Обычно это занимает (1-2 минуты).\n\n") +
		fmt.Sprintf("Сумма вывода: <b>%s NEAR</b>.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Адрес <b>%s</b>.\n\n", wallet) +
		fmt.Sprintf("Комиссия: <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func WithdrawIncorrectWithdrawalAddress() string {
	return "<b>Некорректный адрес</b>\n\n" +
		fmt.Sprintf("Похоже, адрес, который вы прислали некорректен. Пожалуйста, пришлите корректный адрес.\n") +
		fmt.Sprintf("Например: textmoneybot.near\n\n") +
		fmt.Sprintf("Если вы уверены, что адрес корректный, но продолжаете видеть эту ошибку, пожалуйста, обратитесь в службу поддержки.")
}

func WithdrawUnexpectedError() string {
	return "<b>Неожиданная ошибка</b>\n\n" +
		"К сожалению, при обработке вашего запроса произошла ошибка. Команда разработки уведомлена и уже занимается исправлением.\n" +
		"Попробуйте создать вывод еще раз или обратитесь в службу поддержки, если видете это сообщение не в первый раз."
}

func calculateMaxWithdrawAmount(balance int) float64 {
	return utils.GetMaxWithdrawAmount(balance)
}

func TransferAccepted(amount uint64, transferID string) string {
	return "<b>Перевод получен</b>\n\n" +
		fmt.Sprintf("Получатель принял перевод на сумму <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("<i>ID перевода: %s</i>", transferID)
}

func TransferCanceled(amount uint64, transferID string) string {
	return "<b>Перевод отменен</b>\n\n" +
		fmt.Sprintf("Отправитель отменил перевод на сумму <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("<i>ID перевода: %s</i>", transferID)
}

func TransferRejected(amount uint64, transferID string) string {
	return "<b>Перевод отклонен</b>\n\n" +
		fmt.Sprintf("Получатель отклонил перевод на сумму <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("<i>ID перевода: %s</i>", transferID)
}

func Support() string {
	return "<b>TextMoney • поддержка</b>\n\n" +
		"Если у вас возник вопрос по работе бота, проверьте раздел Часто задаваемых вопросов или напишите нам.\n\n" +
		"<b>Обратите внимание:</b> официальный аккаунт поддержки @textmoney_support. Если вам пишут с другого аккаунта - это мошенники.\n\n" +
		"Служба поддержки <b>не просит</b> перевести деньги."
}

func UnsupportedMessage() string {
	return "<b>Ошибка</b>\n\n" +
		"Бот не может обработать ваше сообщение. Если вы уверены, что это ошибка, обратитесь в поддержку."
}

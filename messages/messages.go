package messages

import (
	"fmt"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/utils"
	"math"
	"time"
)

func Welcome(balance int) string {
	return "<b>Добро пожаловать!</b>\n\n" +
		"С помощью этого бота вы можете перевести деньги любому пользователю Telegram," +
		" просто отправив сообщение с суммой.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>", float64(balance/100))
}

func Balance(balance int) string {
	return "<b>TextMoney</b>\n\n" +
		"С помощью этого бота вы можете перевести деньги любому пользователю Telegram," +
		" просто отправив сообщение с суммой.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100)) +
		fmt.Sprintf("<i>Обновлено: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func Deposit(balance int) string {
	return "<b>Пополнение баланса.</b>\n\n" +
		"Выберите сумму, которую хотите внести на баланс.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100)) +
		fmt.Sprintf("<i>Обновлено: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func DepositAmount(amount uint64, paymentLink string, paymentID string) string {
	return "<b>Пополнение баланса.</b>\n\n" +
		fmt.Sprintf("Ваш баланс будет пополнен на <b>%d рублей</b>. Продолжить?\n\n", amount) +
		fmt.Sprintf("<b>Комиссия:</b> %d₽\n<b>Сумма с учетом комиссии:</b> %d₽\n\n", uint64(math.Floor(float64(amount)*config.Fee)), uint64(math.Floor(float64(amount)+float64(amount)*config.Fee))) +
		fmt.Sprintf("Ссылка для оплаты: %s\n\n", paymentLink) +
		fmt.Sprintf("<i>ID платежа: %s</i>", paymentID)
}

func DepositProcessed(balance int, paymentID string) string {
	return "<b>Платеж успешно зачислен.</b>\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100)) +
		fmt.Sprintf("<i>Обновлено: %s</i>\n\n", time.Now().Format("2006-01-02 3:04PM")) +
		fmt.Sprintf("<i>ID платежа: %s</i>", paymentID)
}

func Withdraw(balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		fmt.Sprintf("Укажите сумму вывода. Обратите внимание, что комиссия составит %.1f%% от суммы вывода.\n\n", config.Fee*100) +
		fmt.Sprintf("<b>Минимальная сумма:</b> %d₽\n", config.MinWithdrawAmount/100) +
		fmt.Sprintf("<b>Максимальная сумма:</b> %d₽\n\n", calculateMaxWithdrawAmount(balance)) +
		"В настоящее время вывод денег доступен только на банковские карты.\n\n" +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100))
}

func WithdrawLowBalance(balance int) string {
	return "<b>Вывод недоступен</b>\n\n" +
		"На вашем балансе недостаточно денег для вывода.\n" +
		fmt.Sprintf("<b>Минимальная сумма:</b> %d₽\n\n", config.MinWithdrawAmount/100) +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100))
}

func WithdrawIncorrectAmount(balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		"Сумма должна быть целым числом без лишних знаков. Например: 100.\nДля отмены отправьте текст \"отмена\"\n\n" +
		fmt.Sprintf("<b>Минимальная сумма:</b> %d₽\n", config.MinWithdrawAmount/100) +
		fmt.Sprintf("<b>Максимальная сумма:</b> %d₽\n\n", calculateMaxWithdrawAmount(balance)) +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100))
}

func WithdrawConfirmAmount(amount int, balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		fmt.Sprintf("Вы собираетесь вывести <b>%d₽</b>. Пришлите адрес карты, на которую хотите вывести деньги.\n", amount) +
		fmt.Sprintf("Например: 4200 0000 0000 0000\n\n") +
		fmt.Sprintf("<b>Комиссия:</b> %d₽\n", int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("С вашего баланса будет списано <b>%d₽</b> (с учетом комиссии)\n\n", amount+int(math.Round(float64(amount)*config.Fee))) +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100))
}

func WithdrawConfirmFinal(amount int, cardNumber string, balance int) string {
	return "<b>Вывод денег</b>\n\n" +
		fmt.Sprintf("Вы собираетесь вывести <b>%d₽</b>.\n", amount) +
		fmt.Sprintf("Номер карты <b>%s</b>.\n\n", maskCard(cardNumber)) +
		fmt.Sprintf("<b>Комиссия:</b> %d₽\n", int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("Продолжить? С вашего баланса будет списано <b>%d₽</b> (с учетом комиссии)\n\n", amount+int(math.Round(float64(amount)*config.Fee))) +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100))
}

func WithdrawHasPending(amount int, cardNumber string, withdrawalID string) string {
	return "<b>Активный вывод денег</b>\n\n" +
		fmt.Sprintf("Прямо сейчас у вас есть один необработанный запрос на вывод. Чтобы создать новый, отмените текущий или дождитесь окончания обработки. Обычно это занимает (1-2 рабочих дня).\n\n") +
		fmt.Sprintf("Статус: <b>ожидает обработки</b>\n") +
		fmt.Sprintf("Сумма вывода: <b>%d₽</b>.\n", amount) +
		fmt.Sprintf("Номер карты <b>%s</b>.\n\n", maskCard(cardNumber)) +
		fmt.Sprintf("Комиссия: <b>%d₽</b>\n\n", int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func WithdrawHasProcessing(amount int, cardNumber string, withdrawalID string) string {
	return "<b>Активный вывод денег</b>\n\n" +
		fmt.Sprintf("Прямо сейчас у вас есть один запрос на вывод, находящийся в обработке. К сожалению, отменить вывод находящийся в обработке невозможно.\n\n") +
		fmt.Sprintf("Статус: <b>в обработке</b>\n") +
		fmt.Sprintf("Сумма вывода: <b>%d₽</b>.\n", amount) +
		fmt.Sprintf("Номер карты <b>%s</b>.\n\n", maskCard(cardNumber)) +
		fmt.Sprintf("Комиссия: <b>%d₽</b>\n\n", int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func WithdrawCreated(amount int, cardNumber string, withdrawalID string) string {
	return "<b>Запрос на вывод денег создан</b>\n\n" +
		fmt.Sprintf("Обратите внимание, вы можете создать только один запрос на вывод за раз. Чтобы создать новый, отмените текущий или дождитесь окончания обработки. Обычно это занимает (1-2 рабочих дня).\n\n") +
		fmt.Sprintf("Сумма вывода: <b>%d₽</b>.\n", amount) +
		fmt.Sprintf("Номер карты <b>%s</b>.\n\n", maskCard(cardNumber)) +
		fmt.Sprintf("Комиссия: <b>%d₽</b>\n\n", int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func WithdrawCancel(amount int, cardNumber string, withdrawalID string) string {
	return "<b>Подтвердите отмену</b>\n\n" +
		fmt.Sprintf("Вы действительно хотите отменить запрос на вывод <b>%d₽</b> на карту <b>%s</b>?\n\n", amount, maskCard(cardNumber)) +
		fmt.Sprintf("На ваш баланс будет зачислено <b>%d₽</b>.\n\n", amount+int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func WithdrawCanceled(balance int, amount int, cardNumber string, withdrawalID string) string {
	return "<b>Вывод отменен</b>\n\n" +
		fmt.Sprintf("Вы отменили запрос на вывод <b>%d₽</b> на карту <b>%s</b>?\n\n", amount, maskCard(cardNumber)) +
		fmt.Sprintf("На ваш баланс зачислено <b>%d₽</b>.\n\n", amount+int(math.Floor(float64(amount)*config.Fee))) +
		fmt.Sprintf("Текущий баланс: <b>%.2f₽</b>\n\n", float64(balance/100)) +
		fmt.Sprintf("<i>ID вывода: %s</i>", withdrawalID)
}

func maskCard(cardNumber string) string {
	return fmt.Sprintf("%s %s %s %s", cardNumber[:4], "****", "****", cardNumber[len(cardNumber)-4:])
}

func WithdrawIncorrectCardNumber() string {
	return "<b>Некорректный номер карты</b>\n\n" +
		fmt.Sprintf("Похоже, номер карты, который вы прислали некорректен. Пожалуйста, пришлите корректный номер карты.\n") +
		fmt.Sprintf("Например: 4200 0000 0000 0000\n\n") +
		fmt.Sprintf("Если вы уверены, что номер корректный, но продолжаете видеть эту ошибку, пожалуйста, обратитесь в службу поддержки.")
}

func WithdrawUnexpectedError() string {
	return "<b>Неожиданная ошибка</b>\n\n" +
		"К сожалению, при обработке вашего запроса произошла ошибка. Команда разработки уведомлена и уже занимается исправлением.\n" +
		"Попробуйте создать вывод еще раз или обратитесь в службу поддержки, если видете это сообщение не в первый раз."
}

func calculateMaxWithdrawAmount(balance int) int {
	return utils.GetMaxWithdrawAmount(balance)
}

func TransferAccepted(amount uint64, transferID string) string {
	return "<b>Перевод получен</b>\n\n" +
		fmt.Sprintf("Получатель принял перевод на сумму <b>%.2f₽</b>\n\n", float64(amount/100)) +
		fmt.Sprintf("<i>ID перевода: %s</i>", transferID)
}

func TransferCanceled(amount uint64, transferID string) string {
	return "<b>Перевод отменен</b>\n\n" +
		fmt.Sprintf("Отправитель отменил перевод на сумму <b>%.2f₽</b>\n\n", float64(amount/100)) +
		fmt.Sprintf("<i>ID перевода: %s</i>", transferID)
}

func TransferRejected(amount uint64, transferID string) string {
	return "<b>Перевод отклонен</b>\n\n" +
		fmt.Sprintf("Получатель отклонил перевод на сумму <b>%.2f₽</b>\n\n", float64(amount/100)) +
		fmt.Sprintf("<i>ID перевода: %s</i>", transferID)
}

func UnsupportedMessage() string {
	return "<b>Ошибка</b>\n\n" +
		"Бот не может обработать ваше сообщение. Если вы уверены, что это ошибка, обратитесь в поддержку."
}

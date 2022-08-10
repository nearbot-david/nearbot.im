package messages

import (
	"fmt"
	"github.com/Pay-With-NEAR/nearbot.im/config"
	"github.com/Pay-With-NEAR/nearbot.im/utils"
	"time"
)

func Welcome(balance int) string {
	return "<b>Welcome, my friend!</b>\n\n" +
		"This bot makes it possible to send NEAR to any Telegram user even they don't have NEAR wallet.\n\n" +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>", utils.DisplayAmount(balance))
}

func IncomingTransfer(deposit int, address string, balance int, hash string) string {
	return "<b>NEAR added.</b>\n\n" +
		fmt.Sprintf("You received <b>%s NEAR</b> from <b>%s</b> via on-chain transfer.\n\n", utils.DisplayAmount(deposit), address) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Updated: %s</i>\n\n", time.Now().Format("2006-01-02 3:04PM")) +
		fmt.Sprintf("<i>tx hash: %s</i>", hash)
}

func Balance(balance int) string {
	return "<b>Send NEAR</b>\n\n" +
		"This bot makes it possible to send NEAR to any Telegram user even they don't have NEAR wallet.\n\n" +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Updated: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func Address(address string, balance int) string {
	return "<b>Send NEAR</b>\n\n" +
		fmt.Sprintf("Your NEAR address: <code>%s</code><a href=\"https://nearbot.im/address/%s\">\n</a>\n", address, address) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Updated: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func Deposit(balance int) string {
	return "<b>Top up.</b>\n\n" +
		"Choose how much you want to add to your balance.\n\n" +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Updated: %s</i>", time.Now().Format("2006-01-02 3:04PM"))
}

func DepositAmount(amount float64, paymentLink string, paymentID string) string {
	return "<b>Top up.</b>\n\n" +
		fmt.Sprintf("Amount <b>%s NEAR</b> will be added to your balance. Continue?\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("Payment link: %s\n\n", paymentLink) +
		fmt.Sprintf("<i>Payment ID: %s</i>", paymentID)
}

func DepositProcessed(balance int, paymentID string) string {
	return "<b>NEAR added.</b>\n\n" +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance)) +
		fmt.Sprintf("<i>Updated: %s</i>\n\n", time.Now().Format("2006-01-02 3:04PM")) +
		fmt.Sprintf("<i>Payment ID: %s</i>", paymentID)
}

func Withdraw(balance int) string {
	return "<b>Transfer NEAR</b>\n\n" +
		fmt.Sprintf("How much do you want to transfer? Please note that the fee will be %.1f%% of the amount.\n\n", config.Fee*100) +
		fmt.Sprintf("<b>Min amount:</b> %.5f NEAR\n", utils.GetMinWithdrawAmount()) +
		fmt.Sprintf("<b>Max amount:</b> %.5f NEAR\n\n", calculateMaxWithdrawAmount(balance)) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawLowBalance(balance int) string {
	return "<b>Not enough NEAR</b>\n\n" +
		"You don't have enough NEAR to transfer.\n" +
		fmt.Sprintf("<b>Min amount:</b> %.5f NEAR\n\n", utils.GetMinWithdrawAmount()) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawIncorrectAmount(balance int) string {
	return "<b>Transfer NEAR</b>\n\n" +
		"How much do you want to transfer? For example 0.2\nIf you want to abort transfer, just send \"cancel\" in response.\n\n" +
		fmt.Sprintf("<b>Min amount:</b> %.5f NEAR\n", utils.GetMinWithdrawAmount()) +
		fmt.Sprintf("<b>Max amount:</b> %.5f NEAR\n\n", calculateMaxWithdrawAmount(balance)) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawConfirmAmount(amount int, balance int) string {
	return "<b>Transfer NEAR</b>\n\n" +
		fmt.Sprintf("You are going to transfer <b>%s NEAR</b>.\n\n", utils.DisplayAmount(amount)) +
		"Please send us the address of the wallet to which you want to send the NEAR.\n" +
		fmt.Sprintf("For example: textmoneybot.near\n\n") +
		fmt.Sprintf("<b>Fee:</b> %s NEAR\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("<b>%s NEAR</b> will be deducted from your balance (fee inc.)\n\n", utils.DisplayAmount(int(float64(amount)*(1+config.Fee)))) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawConfirmFinal(amount int, wallet string, balance int) string {
	return "<b>Transfer NEAR</b>\n\n" +
		fmt.Sprintf("You are going to transfer <b>%s NEAR</b>.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Recepient: <b>%s</b>.\n\n", wallet) +
		fmt.Sprintf("<b>Fee:</b> %s NEAR\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("<b>%s NEAR</b> will be deducted from your balance (fee inc.). Continue? \n\n", utils.DisplayAmount(int(float64(amount)*(1+config.Fee)))) +
		fmt.Sprintf("Current balance: <b>%s NEAR</b>\n\n", utils.DisplayAmount(balance))
}

func WithdrawHasProcessing(amount int, wallet string, withdrawalID string) string {
	return "<b>You have active transfer</b>\n\n" +
		fmt.Sprintf("At the moment you already have an active transfer.\n\n") +
		fmt.Sprintf("Status: <b>processing</b>\n") +
		fmt.Sprintf("Amount: <b>%s NEAR</b>.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Recepient: <b>%s</b>.\n\n", wallet) +
		fmt.Sprintf("Fee: <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>Transfer ID: %s</i>", withdrawalID)
}

func WithdrawCreated(amount int, wallet string, withdrawalID string) string {
	return "<b>Transfer sent</b>\n\n" +
		fmt.Sprintf("Note that you can only send one transfer at a time. To create a new one, wait until processing is complete. This usually takes (1-2 minutes).\n\n") +
		fmt.Sprintf("Amount: <b>%s NEAR</b>.\n", utils.DisplayAmount(amount)) +
		fmt.Sprintf("Recepient: <b>%s</b>.\n\n", wallet) +
		fmt.Sprintf("Fee: <b>%s NEAR</b>\n\n", utils.DisplayAmount(int(float64(amount)*config.Fee))) +
		fmt.Sprintf("<i>Transfer ID: %s</i>", withdrawalID)
}

func WithdrawIncorrectWithdrawalAddress() string {
	return "<b>Incorrect wallet address</b>\n\n" +
		fmt.Sprintf("The wallet you sent seems to be incorrect. Please send the correct address.\n") +
		fmt.Sprintf("For example: textmoneybot.near\n\n") +
		fmt.Sprintf("If you are sure that the address is correct, but keep seeing this error, please contact support.")
}

func WithdrawUnexpectedError() string {
	return "<b>Unexpected error</b>\n\n" +
		"Unfortunately, there was an error in processing your request. The development team has been notified and is already working on a fix.\n" +
		"Try to transfer again or contact support if this is not the first time you have seen this message."
}

func calculateMaxWithdrawAmount(balance int) float64 {
	return utils.GetMaxWithdrawAmount(balance)
}

func TransferAccepted(amount uint64, transferID string) string {
	return "<b>Transfer accepted</b>\n\n" +
		fmt.Sprintf("Receiver accepted <b>%s NEAR</b> transfer\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("<i>Transfer ID: %s</i>", transferID)
}

func TransferCanceled(amount uint64, transferID string) string {
	return "<b>Transfer canceled</b>\n\n" +
		fmt.Sprintf("Sender canceled <b>%s NEAR</b> transfer\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("<i>Transfer ID: %s</i>", transferID)
}

func TransferRejected(amount uint64, transferID string) string {
	return "<b>Transfer rejected</b>\n\n" +
		fmt.Sprintf("Receiver declined <b>%s NEAR</b> transfer\n\n", utils.DisplayAmount(int(amount))) +
		fmt.Sprintf("<i>Transfer ID: %s</i>", transferID)
}

func Support() string {
	return "<b>Send NEAR • Support</b>\n\n" +
		"If you have a question or problem, please check Frequently Asked Questions or write to us\n\n" +
		"<b>NOTE:</b> official support account is @textmoney_support. If the other account appears to be TextMoney Support, it's a scammer.\n\n" +
		"Support <b>never</b> asks to transfer your NEAR to someone else."
}

func UnsupportedMessage() string {
	return "<b>Error</b>\n\n" +
		"Bot cannot process your message. If you sure that this is an error please contact support."
}

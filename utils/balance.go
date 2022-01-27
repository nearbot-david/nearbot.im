package utils

import (
	"github.com/mazanax/moneybot/config"
	"math"
)

func GetMaxWithdrawAmount(balance int) int {
	return int(math.Min(math.Ceil(float64(balance)/(1+config.Fee))/100, config.MaxWithdrawAmount/100))
}

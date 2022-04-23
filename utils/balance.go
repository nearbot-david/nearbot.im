package utils

import (
	"fmt"
	"github.com/mazanax/moneybot/config"
	"math"
)

func GetMinWithdrawAmount() float64 {
	return config.MinWithdrawAmount / 1e5
}

func GetMaxWithdrawAmount(balance int) float64 {
	return math.Min(float64(balance)/(1+config.Fee)/1e5, config.MaxWithdrawAmount/1e5)
}

func DisplayAmount(balance int) string {
	return fmt.Sprintf("%.5f", float64(balance)/1e5)
}

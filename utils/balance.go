package utils

import (
	"fmt"
	"github.com/Pay-With-NEAR/nearbot.im/config"
	"math"
	"strings"
)

func GetMinWithdrawAmount() float64 {
	return config.MinWithdrawAmount / 1e5
}

func GetMaxWithdrawAmount(balance int) float64 {
	return math.Min(float64(balance)/(1+config.Fee)/1e5, config.MaxWithdrawAmount/1e5)
}

func DisplayAmount(balance int) string {
	output := fmt.Sprintf("%.5f", float64(balance)/1e5)
	if !strings.Contains(output, ".") {
		return output
	}

	output = strings.TrimRight(output, "0")
	if output[len(output)-1:] == "." {
		output += "0"
	}

	return output
}

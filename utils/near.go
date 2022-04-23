package utils

import (
	"fmt"
	"github.com/mazanax/moneybot/config"
	"os/exec"
	"sync"
)

func IsNearWalletValid(wallet string) bool {
	cmd := exec.Command("near", "state", wallet)
	cmd.Env = []string{"NODE_ENV", config.NearNetwork}
	_, err := cmd.Output()
	if err != nil {
		return false
	}

	return true
}

func Withdraw(wallet string, amount uint64, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command("near", "send", config.NearWallet, wallet, fmt.Sprintf("%.5f", float64(amount)/1e5))
	cmd.Env = []string{"NODE_ENV", config.NearNetwork}
	err := cmd.Run()
	if err != nil {
		return
	}
}

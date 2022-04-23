package config

import "os"

const Fee = 0.05
const MinWithdrawAmount = 5e4    // 0.5 NEAR
const MaxWithdrawAmount = 1000e5 // 1000 NEAR

var NearWallet = os.Getenv("NEAR_WALLET")
var NearNetwork = os.Getenv("NEAR_NETWORK")

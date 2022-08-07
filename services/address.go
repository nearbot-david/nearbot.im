package services

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/repository"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

type AddressManager struct {
	repo *repository.BalanceRepository
}

func NewAddressManager(balanceRepository *repository.BalanceRepository) *AddressManager {
	return &AddressManager{repo: balanceRepository}
}

type NearCredentials struct {
	AccountId  string `json:"account_id"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func (am *AddressManager) GenerateAddress() (string, error) {
	uuid_ := uuid.New()

	cmd := exec.Command("near", "generate-key", uuid_.String())
	cmd.Env = []string{"NEAR_ENV=" + config.NearNetwork}
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Err: %s, output: %s, env: %s\n", err, output, config.NearNetwork)
		return "", err
	}

	keyPath := path.Join(config.NearCreddir, config.NearNetwork, uuid_.String()+".json")
	if _, err := os.Stat(keyPath); err != nil {
		fmt.Printf("Key file not found in %s\n", keyPath)
		return "", err
	}

	credentials, err := parseCredentialsFile(keyPath)
	if err != nil {
		fmt.Printf("Cannot parse credentials file %s: %s\n", keyPath, err)
		return "", err
	}

	publicKeyBytes := base58.Decode(strings.Split(credentials.PublicKey, ":")[1])
	publicKey := hex.EncodeToString(publicKeyBytes)

	err = os.Rename(keyPath, path.Join(config.NearCreddir, config.NearNetwork, publicKey+".json"))
	if err != nil {
		fmt.Printf("Cannot rename credentials file %s: %s\n", keyPath, err)
		return "", err
	}

	return publicKey, nil
}

func (am *AddressManager) Transfer(from string, to string, amount uint64, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command("near", "send", from, to, fmt.Sprintf("%.5f", float64(amount)/1e5))
	cmd.Env = []string{"NODE_ENV=" + config.NearNetwork}
	err := cmd.Run()
	if err != nil {
		return
	}
}

func parseCredentialsFile(path string) (NearCredentials, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return NearCredentials{}, err
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			fmt.Printf("Cannot close file: %s\n", err)
		}
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return NearCredentials{}, err
	}

	var credentials NearCredentials
	err = json.Unmarshal(byteValue, &credentials)
	if err != nil {
		return NearCredentials{}, err
	}

	return credentials, nil
}

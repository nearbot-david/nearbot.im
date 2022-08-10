package services

import (
	"encoding/json"
	"fmt"
	"github.com/Pay-With-NEAR/nearbot.im/config"
	"github.com/Pay-With-NEAR/nearbot.im/messages"
	"github.com/Pay-With-NEAR/nearbot.im/models"
	"github.com/Pay-With-NEAR/nearbot.im/repository"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exec"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"math/big"
	"net/http"
	"oya.to/namedlocker"
	"sync"
	"time"
)

type TransactionChecker struct {
	db           *goqu.Database
	repo         *repository.BalanceRepository
	transactions *repository.TransactionRepository
	sto          *namedlocker.Store
}

func NewTransactionChecker(db *goqu.Database, repo *repository.BalanceRepository, transactions *repository.TransactionRepository) *TransactionChecker {
	return &TransactionChecker{
		db:           db,
		repo:         repo,
		transactions: transactions,
		sto:          &namedlocker.Store{},
	}
}

type OnChainTx struct {
	BlockHash      string      `json:"block_hash"`
	BlockTimestamp string      `json:"block_timestamp"`
	Hash           string      `json:"hash"`
	ActionIndex    int         `json:"action_index"`
	SignerId       string      `json:"signer_id"`
	ReceiverId     string      `json:"receiver_id"`
	ActionKind     string      `json:"action_kind"`
	Args           OnChainArgs `json:"args"`
}

type OnChainArgs struct {
	Deposit string `json:"deposit"`
}

func (tc *TransactionChecker) Run(bot *tg.BotAPI) {
	for {
		wg := &sync.WaitGroup{}
		func(wg *sync.WaitGroup) {
			defer time.Sleep(30 * time.Second)

			scanner, err := tc.db.From("balance").Select("near_address").Where(goqu.C("near_address").Neq("")).Executor().Scanner()
			if err != nil {
				fmt.Printf("Cannot get addresses: %s\n", err)
				return
			}
			defer func(scanner exec.Scanner) {
				err := scanner.Close()
				if err != nil {
					fmt.Printf("Cannot close scanner: %s\n", err)
				}
			}(scanner)

			for scanner.Next() {
				var address string
				err := scanner.ScanVal(&address)
				if err != nil {
					continue
				}

				wg.Add(1)
				go tc.Process(bot, address, wg)
			}
		}(wg)

		wg.Wait()
	}
}

func (tc *TransactionChecker) Process(bot *tg.BotAPI, address string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Processing %s\n", address)

	activity, err := requestAddressActivity(address)
	if err != nil {
		return
	}

	for _, tx := range activity {
		tc.sto.Lock(tx.Hash)

		func(tc *TransactionChecker, tx *OnChainTx, wg *sync.WaitGroup) {
			defer tc.sto.Unlock(tx.Hash)

			if tx.ActionKind != "TRANSFER" {
				return
			}
			if tx.SignerId == config.NearWallet {
				return
			}
			if tx.SignerId == address {
				return
			}
			if tc.repo.FindByAddress(tx.SignerId) != nil {
				return // signer is internal bot wallet
			}
			if tc.transactions.FindByHash(tx.Hash) != nil {
				return // already processed
			}

			amount := new(big.Int)
			amount, ok := amount.SetString(tx.Args.Deposit, 10)
			if !ok {
				fmt.Printf("Cannot parse deposit: %s\n", tx.Args.Deposit)
				return
			}
			amountFloat := new(big.Float).SetInt(amount)

			divider := new(big.Float)
			divider, _ = divider.SetString("1000000000000000000000000")
			amountFloat = amountFloat.Quo(amountFloat, divider)
			amountFloat32, _ := amountFloat.Float32()
			amountInt := int(amountFloat32 * 1e5)
			if amountInt == 0 {
				return
			}

			balance := tc.repo.FindByAddress(tx.ReceiverId)
			balance.NearAmount += uint64(amountInt)
			tc.repo.Persist(balance)

			message := tg.NewMessage(balance.TelegramID, messages.IncomingTransfer(amountInt, tx.SignerId, int(balance.NearAmount), tx.Hash))
			message.ParseMode = tg.ModeHTML
			bot.Send(message)

			transaction := models.Transaction{
				Hash:      tx.Hash,
				Address:   tx.ReceiverId,
				Amount:    uint64(amountInt),
				CreatedAt: time.Now(),
			}
			tc.transactions.Persist(&transaction)

			wg.Add(1)
			go func() {
				time.Sleep(time.Second / 2)

				message := tg.NewMessage(
					balance.TelegramID,
					messages.Balance(int(balance.NearAmount)),
				)
				message.ParseMode = tg.ModeHTML
				message.ReplyMarkup = getBalanceKeyboard()
				bot.Send(message)
				wg.Done()
			}()

			wg.Wait()
		}(tc, &tx, wg)
	}
}

func requestAddressActivity(address string) ([]OnChainTx, error) {
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.kitwallet.app/account/%s/activity", address), nil)
	if err != nil {
		fmt.Printf("HTTP ERR: %s\n", err)
		return []OnChainTx{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP ERR: %s\n", err)
		return []OnChainTx{}, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Printf("HTTP PARSE ERR: %s\n", err)
		return []OnChainTx{}, err
	}

	txData := make([]OnChainTx, 0)
	if err := json.Unmarshal(body, &txData); err != nil {
		fmt.Printf("HTTP JSON PARSE ERR: %s\n", err)
		return []OnChainTx{}, err
	}

	return txData, nil
}

func getBalanceKeyboard() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Balance", "show_balance"),
			tg.NewInlineKeyboardButtonData("Address", "show_address"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Top up", "deposit"),
			tg.NewInlineKeyboardButtonData("Transfer", "withdraw"),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("History", "history"),
		),
	)
}

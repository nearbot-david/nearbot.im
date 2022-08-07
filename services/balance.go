package services

import (
	"github.com/mazanax/moneybot/config"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/utils"
	"log"
	"sync"
	"time"
)

type BalanceManager struct {
	balanceRepository  *repository.BalanceRepository
	transferRepository *repository.TransferRepository
	addressManager     *AddressManager
}

func NewBalanceManager(balanceRepository *repository.BalanceRepository, transferRepository *repository.TransferRepository, addressManager *AddressManager) *BalanceManager {
	return &BalanceManager{
		balanceRepository:  balanceRepository,
		transferRepository: transferRepository,
		addressManager:     addressManager,
	}
}

func (bm *BalanceManager) GetCurrentBalance(telegramID int64) int {
	balance := bm.balanceRepository.FindByTelegramID(telegramID)
	if balance == nil {
		address, err := bm.addressManager.GenerateAddress()
		if err != nil {
			log.Println(err)
			address = ""
		}
		if address != "" {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			defer wg.Wait()

			go bm.addressManager.Transfer(config.NearWallet, address, 1e3, wg)
		}

		balance = &models.Balance{
			TelegramID:  telegramID,
			Amount:      0,
			NearAmount:  0,
			NearAddress: address,
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		}

		if err := bm.balanceRepository.Persist(balance); err != nil {
			log.Println(err)
			return 0
		}
	}

	if balance.NearAddress == "" {
		bm.generateAndSaveAddress(balance)
	}

	return int(balance.NearAmount)
}

func (bm *BalanceManager) GetAddressBalance(telegramID int64) (string, int) {
	balance := bm.balanceRepository.FindByTelegramID(telegramID)
	if balance == nil {
		address, err := bm.addressManager.GenerateAddress()
		if err != nil {
			log.Println(err)
			address = ""
		}
		if address != "" {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			defer wg.Wait()

			go bm.addressManager.Transfer(config.NearWallet, address, 1e3, wg)
		}

		balance = &models.Balance{
			TelegramID:  telegramID,
			Amount:      0,
			NearAmount:  0,
			NearAddress: address,
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		}

		if err := bm.balanceRepository.Persist(balance); err != nil {
			log.Println(err)
			return "", 0
		}
	}

	if balance.NearAddress == "" {
		bm.generateAndSaveAddress(balance)
	}

	return balance.NearAddress, int(balance.NearAmount)
}

func (bm *BalanceManager) generateAndSaveAddress(balance *models.Balance) {
	address, err := bm.addressManager.GenerateAddress()
	if err != nil {
		log.Println(err)
		return
	}
	balance.NearAddress = address
	if err = bm.balanceRepository.Persist(balance); err != nil {
		log.Println(err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	go bm.addressManager.Transfer(config.NearWallet, address, 1e3, wg)
}

func (bm *BalanceManager) Increment(telegramID int64, amount uint64) {
	bm.balanceRepository.Increment(telegramID, amount)
}

func (bm *BalanceManager) Decrement(telegramID int64, amount uint64) {
	bm.balanceRepository.Decrement(telegramID, amount)
}

func (bm *BalanceManager) SendMoney(from int64, to int64, amount uint64, messageID string) *models.Transfer {
	bm.balanceRepository.Decrement(from, amount)

	return bm.createTransfer(from, to, amount, messageID)
}

func (bm *BalanceManager) createTransfer(from int64, to int64, amount uint64, messageID string) *models.Transfer {
	for {
		transferID := utils.RandStringBytes(16)
		if nil == bm.transferRepository.FindBySlug(transferID) {
			transfer := &models.Transfer{
				Slug:      transferID,
				From:      from,
				To:        to,
				Amount:    amount,
				Status:    models.TransferStatusPending,
				CreatedAt: time.Now(),
				MessageID: messageID,
			}

			if err := bm.transferRepository.Persist(transfer); err != nil {
				log.Println(err.Error())
				return nil
			}

			return transfer
		}
	}
}

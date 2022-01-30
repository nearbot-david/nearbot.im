package services

import (
	"fmt"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/utils"
	"log"
	"time"
)

type HistoryManager struct {
	repository *repository.HistoryRepository
}

func NewHistoryManager(repository *repository.HistoryRepository) *HistoryManager {
	return &HistoryManager{repository: repository}
}

func (manager *HistoryManager) CreateDeposit(deposit *models.Deposit) {
	historyItem := &models.HistoryItem{
		Type:      models.HistoryTypeDeposit,
		ModelID:   deposit.ID,
		Slug:      deposit.Slug,
		From:      "",
		To:        fmt.Sprintf("%d", deposit.TelegramID),
		Amount:    deposit.Amount,
		Status:    string(models.PaymentStatusSuccess),
		Cause:     "",
		CreatedAt: time.Now(),
	}
	if err := manager.repository.Persist(historyItem); err != nil {
		log.Println(err)
	}
}

func (manager *HistoryManager) CreateTransfer(transfer *models.Transfer) {
	historyItem := &models.HistoryItem{
		Type:      models.HistoryTypeTransfer,
		ModelID:   transfer.ID,
		Slug:      transfer.Slug,
		From:      fmt.Sprintf("%d", transfer.From),
		To:        fmt.Sprintf("%d", transfer.To),
		Amount:    transfer.Amount,
		Status:    string(transfer.Status),
		Cause:     "",
		CreatedAt: time.Now(),
	}
	if err := manager.repository.Persist(historyItem); err != nil {
		log.Println(err)
	}
}

func (manager *HistoryManager) UpdateTransfer(transfer *models.Transfer, reason string) {
	historyItem := manager.repository.FindByTypeAndSlug(models.HistoryTypeTransfer, transfer.Slug)
	if historyItem == nil {
		log.Printf("Cannot find history item with type = %s and slug = %s", models.HistoryTypeTransfer, transfer.Slug)
		return
	}

	historyItem.From = fmt.Sprintf("%d", transfer.From)
	historyItem.To = fmt.Sprintf("%d", transfer.To)
	historyItem.Status = string(transfer.Status)
	historyItem.Cause = reason

	if err := manager.repository.Persist(historyItem); err != nil {
		log.Println(err)
	}
}

func (manager *HistoryManager) CreateWithdrawal(withdrawal *models.Withdrawal) {
	historyItem := &models.HistoryItem{
		Type:      models.HistoryTypeWithdrawal,
		ModelID:   withdrawal.ID,
		Slug:      withdrawal.Slug,
		From:      fmt.Sprintf("%d", withdrawal.TelegramID),
		To:        withdrawal.Address,
		Amount:    withdrawal.Amount,
		Status:    string(withdrawal.Status),
		Cause:     "",
		CreatedAt: time.Now(),
	}
	if err := manager.repository.Persist(historyItem); err != nil {
		log.Println(err)
	}
}

func (manager *HistoryManager) UpdateWithdrawal(withdrawal *models.Withdrawal, reason string) {
	historyItem := manager.repository.FindByTypeAndSlug(models.HistoryTypeWithdrawal, withdrawal.Slug)
	if historyItem == nil {
		log.Printf("Cannot find history item with type = %s and slug = %s", models.HistoryTypeWithdrawal, withdrawal.Slug)
		return
	}

	historyItem.From = fmt.Sprintf("%d", withdrawal.TelegramID)
	historyItem.To = utils.MaskCard(withdrawal.Address)
	historyItem.Status = string(withdrawal.Status)
	historyItem.Cause = reason

	if err := manager.repository.Persist(historyItem); err != nil {
		log.Println(err)
	}
}

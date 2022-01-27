package services

import (
	"fmt"
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"github.com/mazanax/moneybot/utils"
	"log"
	"time"
)

type WithdrawalManager struct {
	repository *repository.WithdrawalRepository
}

var DraftNotFound = fmt.Errorf("draft not found")

func NewWithdrawalManager(repository *repository.WithdrawalRepository) *WithdrawalManager {
	return &WithdrawalManager{
		repository: repository,
	}
}

func (wm *WithdrawalManager) StoreAmount(telegramID int64, amount uint64) error {
	draft := wm.repository.FindDraft(telegramID)
	if draft == nil {
		draft = wm.createDraft(telegramID, amount)
	} else {
		draft.Amount = amount
		return wm.repository.Persist(draft)
	}

	return nil
}

func (wm *WithdrawalManager) GetDraft(telegramID int64) *models.Withdrawal {
	return wm.repository.FindDraft(telegramID)
}

func (wm *WithdrawalManager) StoreAddress(telegramID int64, address string) error {
	draft := wm.repository.FindDraft(telegramID)
	if draft != nil {
		draft.Address = address
		return wm.repository.Persist(draft)
	}

	return DraftNotFound
}

func (wm *WithdrawalManager) GetActiveWithdrawal(telegramID int64) *models.Withdrawal {
	return wm.repository.FindActiveWithdrawal(telegramID)
}

func (wm *WithdrawalManager) ConfirmDraft(draft *models.Withdrawal) error {
	draft.CreatedAt = time.Now()
	draft.UpdatedAt = time.Now()
	draft.Status = models.WithdrawalStatusPending
	return wm.repository.Persist(draft)
}

func (wm *WithdrawalManager) CancelWithdraw(draft *models.Withdrawal) error {
	draft.UpdatedAt = time.Now()
	draft.Status = models.WithdrawalStatusCanceled
	return wm.repository.Persist(draft)
}

func (wm *WithdrawalManager) createDraft(telegramID int64, amount uint64) *models.Withdrawal {
	for {
		withdrawalID := utils.RandStringBytes(16)
		if nil == wm.repository.FindBySlug(withdrawalID) {
			withdrawal := &models.Withdrawal{
				Slug:       withdrawalID,
				TelegramID: telegramID,
				Status:     models.WithdrawalStatusDraft,
				Amount:     amount,
				CreatedAt:  time.Now(),
			}

			if err := wm.repository.Persist(withdrawal); err != nil {
				log.Println(err.Error())
				return nil
			}

			return withdrawal
		}
	}
}

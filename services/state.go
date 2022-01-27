package services

import (
	"github.com/mazanax/moneybot/models"
	"github.com/mazanax/moneybot/repository"
	"log"
)

type StateManager struct {
	repository *repository.StateRepository
}

func NewStateManager(repository *repository.StateRepository) *StateManager {
	return &StateManager{
		repository: repository,
	}
}

func (manager *StateManager) SetState(telegramID int64, state models.UserState, messageID int) {
	entity := manager.repository.FindByTelegramID(telegramID)
	if entity == nil {
		entity = &models.State{
			TelegramID: telegramID,
			State:      state,
			MessageID:  messageID,
		}
	}

	entity.State = state
	entity.MessageID = messageID
	if err := manager.repository.Persist(entity); err != nil {
		log.Println(err)
	}
}

func (manager *StateManager) GetState(telegramID int64) models.UserState {
	state := manager.repository.FindByTelegramID(telegramID)
	if state == nil {
		return models.UserStateIdle
	}

	return state.State
}

func (manager *StateManager) GetPreviousBotMessageID(telegramID int64) int {
	state := manager.repository.FindByTelegramID(telegramID)
	if state == nil {
		return 0
	}

	return state.MessageID
}

package repository

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/mazanax/moneybot/models"
	"log"
	"time"
)

type StateRepository struct {
	table string
	db    *goqu.Database
}

func NewStateRepository(db *goqu.Database) *StateRepository {
	return &StateRepository{
		table: "state",
		db:    db,
	}
}

func (repo *StateRepository) Persist(entity *models.State) error {
	if entity.ID != 0 {
		entity.UpdatedAt = time.Now()
		update := repo.db.
			Update(repo.table).
			Set(*entity).
			Where(goqu.C("id").Eq(entity.ID)).
			Executor()

		if _, err := update.Exec(); err != nil {
			return err
		}

		return nil
	}

	insert := repo.db.
		Insert(repo.table).
		Rows(*entity).
		Returning("id").
		Executor()

	var id int64
	if _, err := insert.ScanVal(&id); err != nil {
		return err
	}

	entity.ID = id
	return nil
}

func (repo *StateRepository) FindByTelegramID(telegramID int64) *models.State {
	var state models.State
	found, err := repo.db.
		From(repo.table).
		Where(goqu.C("telegram_id").Eq(telegramID)).
		ScanStruct(&state)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &state
}

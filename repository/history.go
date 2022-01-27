package repository

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/mazanax/moneybot/models"
	"log"
)

type HistoryRepository struct {
	table string
	db    *goqu.Database
}

func NewHistoryRepository(db *goqu.Database) *HistoryRepository {
	return &HistoryRepository{
		table: "history",
		db:    db,
	}
}

func (repo *HistoryRepository) Persist(entity *models.HistoryItem) error {
	if entity.ID != 0 {
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

func (repo *HistoryRepository) FindByTypeAndSlug(itemType models.HistoryType, slug string) *models.HistoryItem {
	var history models.HistoryItem
	found, err := repo.db.
		From(repo.table).
		Where(goqu.Ex{
			"item_type": itemType,
			"slug":      slug,
		}).
		ScanStruct(&history)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &history
}

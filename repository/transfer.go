package repository

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/mazanax/moneybot/models"
	"log"
	"time"
)

type TransferRepository struct {
	table string
	db    *goqu.Database
}

func NewTransferRepository(db *goqu.Database) *TransferRepository {
	return &TransferRepository{
		table: "transfer",
		db:    db,
	}
}

func (repo *TransferRepository) Persist(entity *models.Transfer) error {
	fmt.Printf("%+v\n", entity)
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

func (repo *TransferRepository) FindBySlug(slug string) *models.Transfer {
	var entity models.Transfer
	found, err := repo.db.
		From(repo.table).
		Where(goqu.C("slug").Eq(slug)).
		ScanStruct(&entity)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &entity
}

func (repo *TransferRepository) GetLastTransfer(from int64) *models.Transfer {
	var entity models.Transfer
	found, err := repo.db.
		From(repo.table).
		Where(goqu.Ex{
			"from":       from,
			"status":     models.TransferStatusPending,
			"created_at": goqu.Op{"gte": time.Now().Add(-5 * time.Second)},
		}).
		Order(goqu.I("created_at").Desc()).
		Limit(1).
		ScanStruct(&entity)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &entity
}

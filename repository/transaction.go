package repository

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/mazanax/moneybot/models"
	"log"
)

type TransactionRepository struct {
	table string
	db    *goqu.Database
}

func NewTransactionRepository(db *goqu.Database) *TransactionRepository {
	return &TransactionRepository{
		table: "transaction",
		db:    db,
	}
}

func (repo *TransactionRepository) Persist(entity *models.Transaction) error {
	fmt.Printf("%+v\n", entity)
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

func (repo *TransactionRepository) FindByHash(hash string) *models.Transaction {
	var entity models.Transaction
	found, err := repo.db.
		From(repo.table).
		Where(goqu.C("hash").Eq(hash)).
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

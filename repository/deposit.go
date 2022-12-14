package repository

import (
	"github.com/Pay-With-NEAR/nearbot.im/models"
	"github.com/doug-martin/goqu/v9"
	"log"
	"time"
)

type DepositRepository struct {
	table string
	db    *goqu.Database
}

func NewDepositRepository(db *goqu.Database) *DepositRepository {
	return &DepositRepository{
		table: "deposit",
		db:    db,
	}
}

func (repo *DepositRepository) Persist(entity *models.Deposit) error {
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

func (repo *DepositRepository) FindBySlug(slug string) *models.Deposit {
	var deposit models.Deposit
	found, err := repo.db.
		From(repo.table).
		Where(goqu.C("slug").Eq(slug)).
		ScanStruct(&deposit)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &deposit
}

func (repo *DepositRepository) FindByExternalID(paymentMethod string, externalID string) *models.Deposit {
	var deposit models.Deposit
	found, err := repo.db.
		From(repo.table).
		Where(goqu.Ex{
			"payment_method": paymentMethod,
			"external_id":    externalID,
		}).
		ScanStruct(&deposit)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &deposit
}

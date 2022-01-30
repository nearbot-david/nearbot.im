package repository

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/mazanax/moneybot/models"
	"log"
	"time"
)

type BalanceRepository struct {
	table string
	db    *goqu.Database
}

func NewBalanceRepository(db *goqu.Database) *BalanceRepository {
	return &BalanceRepository{
		table: "balance",
		db:    db,
	}
}

func (repo *BalanceRepository) Persist(entity *models.Balance) error {
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

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
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

func (repo *BalanceRepository) Increment(telegramID int64, amount uint64) {
	_, err := repo.db.
		Update(repo.table).
		Set(goqu.Record{"amount": goqu.L(fmt.Sprintf("amount + %d", int(amount)))}).
		Where(goqu.C("telegram_id").Eq(telegramID)).
		Executor().
		Exec()

	if err != nil {
		log.Println(err.Error())
	}
}

func (repo *BalanceRepository) Decrement(telegramID int64, amount uint64) {
	_, err := repo.db.
		Update(repo.table).
		Set(goqu.Record{"amount": goqu.L(fmt.Sprintf("amount - %d", int(amount)))}).
		Where(goqu.C("telegram_id").Eq(telegramID)).
		Executor().
		Exec()

	if err != nil {
		log.Println(err.Error())
	}
}

func (repo *BalanceRepository) FindByTelegramID(telegramID int64) *models.Balance {
	var entity models.Balance
	found, err := repo.db.
		From(repo.table).
		Where(goqu.C("telegram_id").Eq(telegramID)).
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

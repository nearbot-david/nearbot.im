package repository

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/mazanax/moneybot/models"
	"log"
	"time"
)

type WithdrawalRepository struct {
	table string
	db    *goqu.Database
}

func NewWithdrawalRepository(db *goqu.Database) *WithdrawalRepository {
	return &WithdrawalRepository{
		table: "withdrawal",
		db:    db,
	}
}

func (repo *WithdrawalRepository) Persist(entity *models.Withdrawal) error {
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

func (repo *WithdrawalRepository) FindBySlug(slug string) *models.Withdrawal {
	var withdrawal models.Withdrawal
	found, err := repo.db.
		From(repo.table).
		Where(goqu.C("slug").Eq(slug)).
		ScanStruct(&withdrawal)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &withdrawal
}

func (repo *WithdrawalRepository) FindActiveWithdrawal(telegramID int64) *models.Withdrawal {
	var withdrawal models.Withdrawal
	found, err := repo.db.
		From(repo.table).
		Where(goqu.Ex{
			"telegram_id": telegramID,
			"status":      goqu.Op{"in": []string{string(models.WithdrawalStatusPending), string(models.WithdrawalStatusProcessing)}},
		}).
		ScanStruct(&withdrawal)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &withdrawal
}

func (repo *WithdrawalRepository) FindDraft(telegramID int64) *models.Withdrawal {
	var withdrawal models.Withdrawal
	found, err := repo.db.
		From(repo.table).
		Where(goqu.Ex{
			"telegram_id": telegramID,
			"status":      models.WithdrawalStatusDraft,
		}).
		ScanStruct(&withdrawal)

	if err != nil {
		log.Println(err)
		return nil
	}
	if !found {
		return nil
	}

	return &withdrawal
}

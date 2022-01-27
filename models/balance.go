package models

import "time"

type Balance struct {
	ID         int64     `db:"id" goqu:"skipinsert"`
	TelegramID int64     `db:"telegram_id"`
	Amount     uint64    `db:"amount"`
	CreatedAt  time.Time `db:"created_at" goqu:"defaultifempty"`
	UpdatedAt  time.Time `db:"updated_at" goqu:"defaultifempty"`
}

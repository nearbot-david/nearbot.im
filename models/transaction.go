package models

import "time"

type Transaction struct {
	ID        int64     `db:"id" goqu:"skipinsert"`
	Hash      string    `db:"hash"`
	Address   string    `db:"address"`
	Amount    uint64    `db:"amount"`
	CreatedAt time.Time `db:"created_at" goqu:"defaultifempty"`
}

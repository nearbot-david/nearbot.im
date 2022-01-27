package models

import "time"

type HistoryType string

const (
	HistoryTypeDeposit    = HistoryType("DEPOSIT")
	HistoryTypeTransfer   = HistoryType("TRANSFER")
	HistoryTypeWithdrawal = HistoryType("WITHDRAWAL")
)

type HistoryItem struct {
	ID        int64       `db:"id" goqu:"skipinsert"`
	Type      HistoryType `db:"item_type"`
	ModelID   int64       `db:"model_id"`
	Slug      string      `db:"slug"`
	From      string      `db:"from"`
	To        string      `db:"to"`
	Amount    uint64      `db:"amount"`
	Status    string      `db:"status"`
	Cause     string      `db:"cause"`
	CreatedAt time.Time   `db:"created_at" goqu:"defaultifempty"`
}

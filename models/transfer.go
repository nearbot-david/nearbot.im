package models

import "time"

type TransferStatus string

const (
	TransferStatusPending  TransferStatus = "PENDING"
	TransferStatusCanceled TransferStatus = "CANCELED"
	TransferStatusAccepted TransferStatus = "ACCEPTED"
	TransferStatusRejected TransferStatus = "REJECTED"
)

type Transfer struct {
	ID        int64          `db:"id" goqu:"skipinsert"`
	Slug      string         `db:"slug"`
	From      int64          `db:"from"`
	To        int64          `db:"to"`
	Amount    uint64         `db:"amount"`
	Status    TransferStatus `db:"status"`
	CreatedAt time.Time      `db:"created_at" goqu:"defaultifempty"`
	UpdatedAt time.Time      `db:"updated_at" goqu:"defaultifempty"`
	MessageID string         `db:"message_id"`
}

package models

import "time"

type WithdrawalStatus string

const (
	WithdrawalStatusDraft      WithdrawalStatus = "DRAFT"
	WithdrawalStatusPending    WithdrawalStatus = "PENDING"
	WithdrawalStatusProcessing WithdrawalStatus = "PROCESSING"
	WithdrawalStatusSuccess    WithdrawalStatus = "SUCCESS"
	WithdrawalStatusRejected   WithdrawalStatus = "REJECTED"
	WithdrawalStatusCanceled   WithdrawalStatus = "CANCELED"
)

type Withdrawal struct {
	ID         int64            `db:"id" goqu:"skipinsert"`
	Slug       string           `db:"slug"`
	TelegramID int64            `db:"telegram_id"`
	Status     WithdrawalStatus `db:"status"`
	Amount     uint64           `db:"amount"`
	Address    string           `db:"address"`
	CreatedAt  time.Time        `db:"created_at" goqu:"defaultifempty"`
	UpdatedAt  time.Time        `db:"updated_at" goqu:"defaultifempty"`
	Comment    string           `db:"comment"`
}

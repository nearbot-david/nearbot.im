package models

import "time"

type PaymentStatus string

const (
	PaymentStatusNew     PaymentStatus = "NEW"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFail    PaymentStatus = "FAIL"
)

type Deposit struct {
	ID         int64         `db:"id" goqu:"skipinsert"`
	Slug       string        `db:"slug"`
	TelegramID int64         `db:"telegram_id"`
	Method     string        `db:"payment_method"`
	Amount     uint64        `db:"amount"`
	Status     PaymentStatus `db:"status"`
	CreatedAt  time.Time     `db:"created_at" goqu:"defaultifempty"`
	UpdatedAt  time.Time     `db:"updated_at" goqu:"defaultifempty"`
	MessageID  int           `db:"message_id" goqu:"defaultifempty"`
}

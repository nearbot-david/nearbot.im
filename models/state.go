package models

import "time"

type UserState string

const (
	UserStateIdle            UserState = "IDLE"
	UserStateWithdrawAmount  UserState = "WITHDRAW_AMOUNT"
	UserStateWithdrawAddress UserState = "WITHDRAW_ADDRESS"
	UserStateWithdrawConfirm UserState = "WITHDRAW_CONFIRM"
)

type State struct {
	ID         int64     `db:"id" goqu:"skipinsert"`
	TelegramID int64     `db:"telegram_id"`
	State      UserState `db:"state"`
	UpdatedAt  time.Time `db:"updated_at" goqu:"defaultifempty"`
	MessageID  int       `db:"message_id" goqu:"defaultifempty"`
}

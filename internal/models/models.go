package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id       uuid.UUID `db:"id"`
	Username string    `json:"login" db:"username"`
	Password string    `json:"password" db:"password"`
}

type UserBalance struct {
	Current   float32 `json:"current" db:"balance"`
	Withdrawn int     `json:"withdrawn" db:"withdrawn"`
}

type UserOrder struct {
	Number     int       `db:"order_number" json:"number"`
	Status     string    `db:"status" json:"status"`
	Accrual    int       `db:"accrual" json:"accrual"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}

type Orders []UserOrder

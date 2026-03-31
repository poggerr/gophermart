package models

import (
	"time"
)

type UserOrder struct {
	Number     string    `db:"order_number" json:"number"`
	Status     string    `db:"status" json:"status"`
	Accrual    *float32  `db:"accrual_service" json:"accrual,omitempty"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}

type Orders []UserOrder

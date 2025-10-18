package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type PaymentModel struct {
	PayID   string           `json:"payment_id"`
	OwnerID string           `json:"owner_id"`
	Status  db.PaymentStatus `json:"status"`
	Type    *string          `json:"type"`
	PayDate *time.Time       `json:"pay_date,omitempty"`
}

package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type PaymentModel struct {
	PayID   string           `json:"payment_id"`
	OwnerID string           `json:"owner_id"`
	Status  db.PaymentStatus `json:"status"`
	Price   int              `json:"price"`
	Type    *string          `json:"type"`
	PayDate *time.Time       `json:"pay_date,omitempty"`
}

type UpdatePaymentRequest struct {
	// ล้อตาม enum payment_status ของคุณ
	Status *string `json:"status" validate:"omitempty,oneof=UNPAID PAID"`

	// สมมติว่า Type มีได้ 2 แบบ (คุณไปแก้ได้)
	Type *string `json:"type" validate:"omitempty,min=1"`

	// validate datetime แบบ RFC3339 (แบบเดียวกับที่โค้ดเก่าคุณพยายาม Parse)
	PayDate *string `json:"pay_date" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type CreatePaymentModel struct {
	ReserveDateStart time.Time `json:"reserve_date_start" validate:"required"`
	ReserveDateEnd   time.Time `json:"reserve_date_end" validate:"required"`
}

type PaymentCommonModel struct {
	Status  db.PaymentStatus `json:"status"`
	Price   int              `json:"price"`
	Type    *string          `json:"type"`
	PayDate *time.Time       `json:"pay_date,omitempty"`
}

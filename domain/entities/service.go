package entities

import (
    "time"
    "lama-backend/domain/prisma/db"
)

type ServiceModel struct {
    Sid        string          `json:"service_id"`
    OwnerID    string          `json:"owner_id"`
    PetID      string          `json:"pet_id"`
    PaymentID  string          `json:"payment_id"`
    Price      int             `json:"price"`
    Status     db.ServiceStatus`json:"status"`
    ReserveDate time.Time      `json:"reserve_date"`
}

type CreateServiceRequest struct {
    OwnerID     string    `json:"owner_id" validate:"required,uuid4"`
    PetID       string    `json:"pet_id" validate:"required,uuid4"`
    PaymentID   string    `json:"payment_id" validate:"required,uuid4"`
    Price       int       `json:"price" validate:"required,gte=0"`
    Status      string    `json:"status" validate:"required,oneof=wait ongoing finish"`
    ReserveDate time.Time `json:"reserve_date" validate:"required"`
}

type UpdateServiceRequest struct {
    OwnerID     *string    `json:"owner_id,omitempty" validate:"omitempty,uuid4"`
    PetID       *string    `json:"pet_id,omitempty" validate:"omitempty,uuid4"`
    PaymentID   *string    `json:"payment_id,omitempty" validate:"omitempty,uuid4"`
    Price       *int       `json:"price,omitempty" validate:"omitempty,gte=0"`
    Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=wait ongoing finish"`
    ReserveDate *time.Time `json:"reserve_date,omitempty"`
}
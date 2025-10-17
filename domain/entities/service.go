package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type ServiceModel struct {
	Sid         string           `json:"service_id"`
	OwnerID     string           `json:"owner_id"`
	PetID       string           `json:"pet_id"`
	PaymentID   string           `json:"payment_id"`
	Price       int              `json:"price"`
	Status      db.ServiceStatus `json:"status"`
	ReserveDate time.Time        `json:"reserve_date"`

	ServiceType string  `json:"service_type"`
	StaffID     string  `json:"staff_id"`
	Disease     *string `json:"disease,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

type CreateServiceRequest struct {
	OwnerID     string    `json:"owner_id" validate:"required,uuid4"`
	PetID       string    `json:"pet_id" validate:"required,uuid4"`
	PaymentID   string    `json:"payment_id" validate:"required,uuid4"`
	StaffID     string    `json:"staff_id" validate:"required,uuid4"`
	ServiceType string    `json:"service_type" validate:"required,oneof=mservice cservice"`
	Price       int       `json:"price" validate:"required,gte=0"`
	Status      string    `json:"status" validate:"required,oneof=wait ongoing finish"`
	ReserveDate time.Time `json:"reserve_date" validate:"required"`
	Disease     *string   `json:"disease,omitempty" validate:"omitempty,min=1"`
	Comment     *string   `json:"comment,omitempty" validate:"omitempty,min=1"`
}

type UpdateServiceRequest struct {
	OwnerID     *string    `json:"owner_id,omitempty" validate:"omitempty,uuid4"`
	PetID       *string    `json:"pet_id,omitempty" validate:"omitempty,uuid4"`
	PaymentID   *string    `json:"payment_id,omitempty" validate:"omitempty,uuid4"`
	StaffID     *string    `json:"staff_id,omitempty" validate:"omitempty,uuid4"`
	ServiceType *string    `json:"service_type,omitempty" validate:"omitempty,oneof=mservice cservice"`
	Price       *int       `json:"price,omitempty" validate:"omitempty,gte=0"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=wait ongoing finish"`
	ReserveDate *time.Time `json:"reserve_date,omitempty"`
	Disease     *string    `json:"disease,omitempty" validate:"omitempty,min=1"`
	Comment     *string    `json:"comment,omitempty" validate:"omitempty,min=1"`
}

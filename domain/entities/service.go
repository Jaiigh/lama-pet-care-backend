package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type ServiceModel struct {
	Sid              string           `json:"service_id"`
	OwnerID          string           `json:"owner_id"`
	PetID            string           `json:"pet_id"`
	PaymentID        string           `json:"payment_id"`
	Price            int              `json:"price"`
	Status           db.ServiceStatus `json:"status"`
	ReserveDateStart time.Time        `json:"reserve_date_start"`
	ReserveDateEnd   time.Time        `json:"reserve_date_end"`

	ServiceType string  `json:"service_type"`
	StaffID     string  `json:"staff_id"`
	Disease     *string `json:"disease,omitempty"`
	Comment     *string `json:"comment,omitempty"`
	Score       *int    `json:"score,omitempty"`
}

type CreateServiceRequest struct {
	OwnerID          string    `json:"owner_id,omitempty" validate:"omitempty,uuid4"`
	PetID            string    `json:"pet_id" validate:"required,uuid4"`
	PaymentID        string    `json:"payment_id,omitempty" validate:"omitempty"` // for backend don't require in request
	StaffID          string    `json:"staff_id" validate:"required,uuid4"`
	ServiceType      string    `json:"service_type" validate:"required,oneof=mservice cservice"`
	Price            int       `json:"price" validate:"required,gte=0"`
	Status           string    `json:"status" validate:"required,oneof=wait ongoing finish"`
	ReserveDateStart time.Time `json:"reserve_date_start" validate:"required"`
	ReserveDateEnd   time.Time `json:"reserve_date_end" validate:"required"`
	Disease          *string   `json:"disease,omitempty" validate:"omitempty,min=1"`
	Comment          *string   `json:"comment,omitempty" validate:"omitempty,min=1"`
}

type UpdateServiceRequest struct {
	OwnerID          *string    `json:"owner_id,omitempty" validate:"omitempty,uuid4"`
	PetID            *string    `json:"pet_id,omitempty" validate:"omitempty,uuid4"`
	StaffID          *string    `json:"staff_id,omitempty" validate:"omitempty,uuid4"`
	Price            *int       `json:"price,omitempty" validate:"omitempty,gte=0"`
	Status           *string    `json:"status,omitempty" validate:"omitempty,oneof=wait ongoing finish"`
	ReserveDateStart *time.Time `json:"reserve_date_start,omitempty"`
	ReserveDateEnd   *time.Time `json:"reserve_date_end,omitempty"`
	Disease          *string    `json:"disease,omitempty" validate:"omitempty,min=1"`
	Comment          *string    `json:"comment,omitempty" validate:"omitempty,min=1"`
	Score            *int       `json:"score,omitempty" validate:"omitempty,gte=1,lte=5"`
}

type SubService struct {
	ServiceID string  `json:"service_id"`
	StaffID   string  `json:"staff_id"`
	Disease   *string `json:"disease,omitempty"`
	Comment   *string `json:"comment,omitempty"`
	Score     *int    `json:"score,omitempty"`
}

type RDateRange struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type ServiceModel struct {
	Sid              string           `json:"service_id"`
	ShowId           int              `json:"show_id"`
	OwnerID          string           `json:"owner_id"`
	PetID            string           `json:"pet_id"`
	PaymentID        string           `json:"payment_id"`
	Status           db.ServiceStatus `json:"status"`
	ReserveDateStart time.Time        `json:"reserve_date_start"`
	ReserveDateEnd   time.Time        `json:"reserve_date_end"`

	ServiceType string          `json:"service_type"`
	StaffID     string          `json:"staff_id"`
	Staff       StaffCommonData `json:"staff"`
	Pet         PetDataModel    `json:"pet"`
	Disease     *string         `json:"disease,omitempty"`
	Comment     *string         `json:"comment,omitempty"`
	Score       *int            `json:"score,omitempty"`
}

type CreateServiceRequest struct {
	OwnerID          string    `json:"owner_id,omitempty" validate:"omitempty,uuid4"` // for admin
	PetID            string    `json:"pet_id" validate:"required,uuid4"`
	PaymentID        string    `json:"payment_id,omitempty"` // for backend
	StaffID          string    `json:"staff_id" validate:"required,uuid4"`
	ServiceType      string    `json:"service_type" validate:"required,oneof=mservice cservice"`
	Status           string    `json:"status" validate:"required,oneof=wait ongoing finish"`
	ReserveDateStart time.Time `json:"reserve_date_start" validate:"required"`
	ReserveDateEnd   time.Time `json:"reserve_date_end" validate:"required"`
}

type UpdateServiceRequest struct {
	OwnerID          *string    `json:"owner_id,omitempty" validate:"omitempty,uuid4"`
	PetID            *string    `json:"pet_id,omitempty" validate:"omitempty,uuid4"`
	StaffID          *string    `json:"staff_id,omitempty" validate:"omitempty,uuid4"`
	Status           *string    `json:"status,omitempty" validate:"omitempty,oneof=wait ongoing finish"`
	ReserveDateStart *time.Time `json:"reserve_date_start,omitempty"`
	ReserveDateEnd   *time.Time `json:"reserve_date_end,omitempty"`
	Disease          *string    `json:"disease,omitempty" validate:"omitempty,min=1"`
	Comment          *string    `json:"comment,omitempty" validate:"omitempty,min=1"`
	Score            *int       `json:"score,omitempty" validate:"omitempty,gte=1,lte=5"`
}

type ReviewRequest struct {
	Comment *string `json:"comment,omitempty" validate:"omitempty,min=1,max=1000"`
	Score   *int    `json:"score,omitempty" validate:"omitempty,gte=1,lte=5"`
}

type ReviewResponse struct {
	ServiceID string  `json:"service_id"`
	StaffID   string  `json:"staff_id"`
	Comment   *string `json:"comment,omitempty"`
	Score     *int    `json:"score,omitempty"`
}

type SubService struct {
	ServiceID string  `json:"service_id"`
	StaffID   string  `json:"staff_id,omitempty"`
	Disease   *string `json:"disease,omitempty"`
	Comment   *string `json:"comment,omitempty"`
	Score     *int    `json:"score,omitempty"`
}

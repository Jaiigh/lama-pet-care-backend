package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type UserDataModel struct {
	UserID          string     `json:"user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Email           string     `json:"email"`
	Password        string     `json:"password"`
	Name            string     `json:"name"`
	BirthDate       time.Time  `json:"birth_date"`
	TelephoneNumber string     `json:"telephone_number"`
	Address         string     `json:"address"`
	LicenseNumber   string     `json:"license_number,omitempty"`
	StartDate       time.Time  `json:"start_date,omitempty"`
	StartWorkTime   time.Time  `json:"start_work_time,omitempty"`
	EndWorkTime     time.Time  `json:"end_work_time,omitempty"`
	Specialization  string     `json:"specialization,omitempty"`
	Rating          db.Decimal `json:"rating,omitempty"`
	TotalSpending   db.Decimal `json:"total_spending,omitempty"`
}

type UserIDModel struct {
	UserID string `json:"user_id"`
}

type LoginUserRequestModel struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginUserResponseModel struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type CreatedUserModel struct {
	Email           string    `json:"email" validate:"required,email"`
	Password        string    `json:"password" validate:"required"`
	Name            string    `json:"name" validate:"required"`
	BirthDate       time.Time `json:"birth_date" validate:"required"`
	TelephoneNumber string    `json:"telephone_number" validate:"required,len=10,numeric"`
	Address         string    `json:"address" validate:"required"`
	LicenseNumber   string    `json:"license_number,omitempty"` // doctor only
	Specialization  string    `json:"specialization,omitempty"` // caretaker only
}

type UpdateUserModel struct {
	Email           *string     `json:"email,omitempty"`
	Password        *string     `json:"password,omitempty"`
	Name            *string     `json:"name,omitempty"`
	BirthDate       *time.Time  `json:"birth_date,omitempty"`
	TelephoneNumber *string     `json:"telephone_number,omitempty"`
	Address         *string     `json:"address,omitempty"`
	LicenseNumber   *string     `json:"license_number,omitempty"`  // doctor only
	StartDate       *time.Time  `json:"start_date,omitempty"`      // doctor only
	StartWorkTime   *time.Time  `json:"start_work_time,omitempty"` // doctor/caretaker only
	EndWorkTime     *time.Time  `json:"end_work_time,omitempty"`   // doctor/caretaker only
	Specialization  *string     `json:"specialization,omitempty"`  // caretaker only
	Rating          *db.Decimal `json:"rating,omitempty"`          // caretaker only
	TotalSpending   *db.Decimal `json:"total_spending,omitempty"`  // owner only
}

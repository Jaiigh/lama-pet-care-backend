package entities

import (
	"time"
)

type UserDataModel struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	Role      string    `json:"role,omitempty"`
}

type UserIDModel struct {
	UserID string `json:"user_id"`
}

type CreatedUserModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

package entities

import (
	"lama-backend/domain/prisma/db"
	"time"
)

type PetDataModel struct {
	PetID     string     `json:"pet_id"`
	OwnerID   string     `json:"owner_id"`
	Breed     string     `json:"breed,omitempty"`
	Name      string     `json:"name,omitempty"`
	BirthDate time.Time  `json:"birth_date"`
	Weight    db.Decimal `json:"weight"`
	Kind      string     `json:"kind"`
	Sex       db.PetSex  `json:"sex"`
}

type CreatedPetModel struct {
	Breed     *string    `json:"breed,omitempty"`
	Name      *string    `json:"name,omitempty"`
	BirthDate time.Time  `json:"birth_date" validate:"required"`
	Weight    db.Decimal `json:"weight" validate:"required"`
	Kind      string     `json:"kind" validate:"required"`
	Sex       db.PetSex  `json:"sex" validate:"required,oneof=male female unknown"`
	OwnerID   string     `json:"owner_id" validate:"required,uuid4"`
}

type UpdatePetModel struct {
	Breed     *string     `json:"breed,omitempty"`
	Name      *string     `json:"name,omitempty"`
	BirthDate *time.Time  `json:"birth_date,omitempty"`
	Weight    *db.Decimal `json:"weight,omitempty"`
	Kind      *string     `json:"kind,omitempty"`
	Sex       *db.PetSex  `json:"sex,omitempty"`
	OwnerID   *string     `json:"owner_id,omitempty" validate:"omitempty,uuid4"`
}

type PetCommonModel struct {
	Breed     string     `json:"breed,omitempty"`
	Name      string     `json:"name,omitempty"`
	BirthDate time.Time  `json:"birth_date"`
	Weight    db.Decimal `json:"weight"`
	Kind      string     `json:"kind"`
	Sex       db.PetSex  `json:"sex"`
}

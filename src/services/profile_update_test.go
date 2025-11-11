// src/services/profile_update_test.go
package services

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"

	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/src/services/mocks"
)

func TestOwnerService_UpdateOwnerByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIOwnerRepository(ctrl)
	svc := &OwnerService{Repo: mockRepo}

	spend := decimal.NewFromInt(9)
	req := entities.UpdateUserModel{
		TotalSpending: &spend,
	}

	mockRepo.EXPECT().
		UpdateByID("owner-1", req).
		Return(&entities.UserDataModel{UserID: "owner-1", TotalSpending: spend}, nil)

	got, err := svc.UpdateOwnerByID("owner-1", req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !got.TotalSpending.Equal(spend) {
		t.Fatalf("want %s got %s", spend, got.TotalSpending)
	}
}

func TestOwnerService_UpdateOwnerByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIOwnerRepository(ctrl)
	svc := &OwnerService{Repo: mockRepo}

	req := entities.UpdateUserModel{}

	mockRepo.EXPECT().
		UpdateByID("owner-1", req).
		Return(nil, errors.New("no fields to update"))

	_, err := svc.UpdateOwnerByID("owner-1", req)
	if err == nil || err.Error() != "no fields to update" {
		t.Fatalf("expected repo err, got %v", err)
	}
}

func TestCaretakerService_UpdateCaretakerByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockICaretakerRepository(ctrl)
	svc := &CaretakerService{Repo: mockRepo}

	spec := "spa"
	start := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(8 * time.Hour)

	req := entities.UpdateUserModel{
		Specialization: &spec,
		StartWorkTime:  &start,
		EndWorkTime:    &end,
	}

	mockRepo.EXPECT().
		UpdateByID("care-1", req).
		Return(&entities.UserDataModel{UserID: "care-1", Specialization: spec}, nil)

	got, err := svc.UpdateCaretakerByID("care-1", req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Specialization != spec {
		t.Fatalf("want spec %s got %s", spec, got.Specialization)
	}
}

func TestDoctorService_UpdateDoctorByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIDoctorRepository(ctrl)
	svc := &DoctorService{Repo: mockRepo}

	license := "LIC-999"
	startDate := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC)

	req := entities.UpdateUserModel{
		LicenseNumber: &license,
		StartDate:     &startDate,
	}

	mockRepo.EXPECT().
		UpdateByID("doc-1", req).
		Return(&entities.UserDataModel{UserID: "doc-1", LicenseNumber: license}, nil)

	got, err := svc.UpdateDoctorByID("doc-1", req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.LicenseNumber != license {
		t.Fatalf("want license %s got %s", license, got.LicenseNumber)
	}
}

func TestUsersService_UpdateUsersByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := mocks.NewMockIUsersRepository(ctrl)
	svc := &UsersService{UsersRepository: mockUsers}

	email := "new@lama.com"
	name := "New Name"
	req := entities.UpdateUserModel{
		Email: &email,
		Name:  &name,
	}

	mockUsers.EXPECT().
		UpdateByID("user-1", req).
		Return(&entities.UserDataModel{UserID: "user-1", Email: email, Name: name}, nil)

	got, err := svc.UpdateUsersByID("user-1", req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Email != email {
		t.Fatalf("want email %s got %s", email, got.Email)
	}
}


func TestPetService_UpdatePet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIPetRepository(ctrl)
	svc := &PetService{PetRepository: mockRepo}

	newName := "Pluto"
	newBreed := "Golden"
	newKind := "dog"
	newSex := db.PetSexMale
	newBirth := time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)
	newWeight := decimal.NewFromFloat(12.5)

	req := entities.UpdatePetModel{
		Name:      &newName,
		Breed:     &newBreed,
		Kind:      &newKind,
		Sex:       &newSex,
		BirthDate: &newBirth,
		Weight:    &newWeight,
	}

	expected := &entities.PetDataModel{
		PetID:     "pet-1",
		OwnerID:   "owner-1",
		Name:      newName,
		Breed:     newBreed,
		Kind:      newKind,
		Sex:       newSex,
		BirthDate: newBirth,
		Weight:    newWeight,
	}

	mockRepo.EXPECT().UpdatePet("pet-1", req).Return(expected, nil)

	got, err := svc.UpdatePet("pet-1", req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Name != newName || got.Breed != newBreed {
		t.Fatalf("want %+v got %+v", expected, got)
	}
}

func TestPetService_UpdatePet_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIPetRepository(ctrl)
	svc := &PetService{PetRepository: mockRepo}

	req := entities.UpdatePetModel{}
	mockRepo.EXPECT().UpdatePet("pet-1", req).Return(nil, errors.New("update failed"))

	if _, err := svc.UpdatePet("pet-1", req); err == nil || err.Error() != "update failed" {
		t.Fatalf("expected repo error, got %v", err)
	}
}
package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type DoctorService struct {
	Repo repositories.IDoctorRepository
}

type IDoctorService interface {
	FindDoctorByID(id string) (*entities.UserDataModel, error)
	DeleteDoctorByID(id string) (*entities.UserDataModel, error)
	UpdateDoctorByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewDoctorService(repo repositories.IDoctorRepository) *DoctorService {
	return &DoctorService{Repo: repo}
}

func (s *DoctorService) FindDoctorByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.Repo.FindByID(id)
}

func (s *DoctorService) DeleteDoctorByID(id string) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.DeleteByID(id)
}

func (s *DoctorService) UpdateDoctorByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.UpdateByID(id, data)
}

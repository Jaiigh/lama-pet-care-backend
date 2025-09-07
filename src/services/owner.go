package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type OwnerService struct {
	Repo repositories.IOwnerRepository
}

type IOwnerService interface {
	FindOwnerByID(id string) (*entities.UserDataModel, error)
	DeleteOwnerByID(id string) (*entities.UserDataModel, error)
	UpdateOwnerByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewOwnerService(repo repositories.IOwnerRepository) IOwnerService {
	return &OwnerService{Repo: repo}
}

func (s *OwnerService) FindOwnerByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.Repo.FindByID(id)
}

func (s *OwnerService) DeleteOwnerByID(id string) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.DeleteByID(id)
}

func (s *OwnerService) UpdateOwnerByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.UpdateByID(id, data)
}

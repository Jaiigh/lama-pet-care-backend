package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type OwnerService struct {
	Repo repositories.IOwnerRepository
}

func NewOwnerService(repo repositories.IOwnerRepository) *OwnerService {
	return &OwnerService{Repo: repo}
}

func (s *OwnerService) GetOwnerByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.Repo.FindByID(id)
}

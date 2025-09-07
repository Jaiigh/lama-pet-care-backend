package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type CaretakerService struct {
	Repo repositories.ICaretakerRepository
}

type ICaretakerService interface {
	FindCaretakerByID(id string) (*entities.UserDataModel, error)
	DeleteCaretakerByID(id string) (*entities.UserDataModel, error)
	UpdateCaretakerByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewCaretakerService(repo repositories.ICaretakerRepository) ICaretakerService {
	return &CaretakerService{Repo: repo}
}

func (s *CaretakerService) FindCaretakerByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.Repo.FindByID(id)
}

func (s *CaretakerService) DeleteCaretakerByID(id string) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.DeleteByID(id)
}

func (s *CaretakerService) UpdateCaretakerByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.UpdateByID(id, data)
}

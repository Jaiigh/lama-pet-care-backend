package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type UsersService struct {
	Repo repositories.IUsersRepository
}

type IUsersService interface {
	FindUsersByID(id string) (*entities.UserDataModel, error)
	DeleteUsersByID(id string) (*entities.UserDataModel, error)
	UpdateUsersByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewUsersService(repo repositories.IUsersRepository) IUsersService {
	return &UsersService{Repo: repo}
}

func (s *UsersService) FindUsersByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.Repo.FindByID(id)
}

func (s *UsersService) DeleteUsersByID(id string) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.DeleteByID(id)
}

func (s *UsersService) UpdateUsersByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.UpdateByID(id, data)
}

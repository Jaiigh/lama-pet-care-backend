package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type UsersService struct {
	UsersRepository     repositories.IUsersRepository
	OwnerRepository     repositories.IOwnerRepository
	CaretakerRepository repositories.ICaretakerRepository
	DoctorRepository    repositories.IDoctorRepository
}

type IUsersService interface {
	FindUsersByID(id string) (*entities.UserDataModel, error)
	DeleteUsersByID(id string) (*entities.UserDataModel, error)
	UpdateUsersByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewUsersService(repoUsers repositories.IUsersRepository, repoOwner repositories.IOwnerRepository, repoCaretaker repositories.ICaretakerRepository, repoDoctor repositories.IDoctorRepository) IUsersService {
	return &UsersService{
		UsersRepository:     repoUsers,
		OwnerRepository:     repoOwner,
		CaretakerRepository: repoCaretaker,
		DoctorRepository:    repoDoctor,
	}
}

func (s *UsersService) FindUsersByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.UsersRepository.FindByID(id)
}

func (s *UsersService) DeleteUsersByID(id string) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.UsersRepository.DeleteByID(id)
}

func (s *UsersService) UpdateUsersByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.UsersRepository.UpdateByID(id, data)
}

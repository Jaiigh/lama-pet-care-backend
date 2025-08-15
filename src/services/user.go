package services

import (
	"fmt"
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
	"lama-backend/src/utils"
	"net/mail"
)

type usersService struct {
	UsersRepository repositories.IUsersRepository
}

type IUsersService interface {
	GetAllUsers() (*[]entities.UserDataModel, error)
	InsertNewUser(data entities.CreatedUserModel) (*entities.UserDataModel, error)
	GetByID(userID string) (*entities.UserDataModel, error)
	UpdateUser(data entities.UserDataModel) (*entities.UserDataModel, error)
	DeleteUser(userID string) error
	Login(data entities.CreatedUserModel) (*entities.UserDataModel, error)
}

func NewUsersService(repo0 repositories.IUsersRepository) IUsersService {
	return &usersService{
		UsersRepository: repo0,
	}
}

func (sv *usersService) GetAllUsers() (*[]entities.UserDataModel, error) {
	data, err := sv.UsersRepository.FindAll()
	if err != nil {
		return nil, err
	}

	return data, nil

}

func (sv *usersService) GetByID(userID string) (*entities.UserDataModel, error) {
	data, err := sv.UsersRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (sv *usersService) InsertNewUser(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	//check email format
	if _, err := mail.ParseAddress(data.Email); err != nil {
		return nil, err
	}

	return sv.UsersRepository.InsertUser(data)
}

func (sv *usersService) UpdateUser(data entities.UserDataModel) (*entities.UserDataModel, error) {
	// Validate email format
	if _, err := mail.ParseAddress(data.Email); err != nil {
		return nil, err
	}

	return sv.UsersRepository.UpdateUser(data)
}

func (sv *usersService) DeleteUser(userID string) error {
	if _, err := sv.UsersRepository.FindByID(userID); err != nil {
		return err
	}

	return sv.UsersRepository.DeleteUser(userID)
}

func (sv *usersService) Login(data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	// Validate email format
	if _, err := mail.ParseAddress(data.Email); err != nil {
		return nil, err
	}

	userData, err := sv.UsersRepository.FindByEmail(data.Email)
	if err != nil {
		return nil, err
	}
	if !utils.CheckPasswordHash(data.Password, userData.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	return userData, nil
}

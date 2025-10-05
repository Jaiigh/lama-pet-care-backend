package services

import (
	"fmt"
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"
)

type authService struct {
	UsersRepository     repositories.IUsersRepository
	OwnerRepository     repositories.IOwnerRepository
	CaretakerRepository repositories.ICaretakerRepository
	DoctorRepository    repositories.IDoctorRepository
}

type IAuthService interface {
	CheckToken(td *middlewares.TokenDetails) error
	Register(role string, data entities.CreatedUserModel) (*entities.UserDataModel, error)
	Login(role string, data entities.LoginUserRequestModel) (*entities.LoginUserResponseModel, error)
}

func NewAuthService(repoUsers repositories.IUsersRepository, repoOwner repositories.IOwnerRepository, repoCaretaker repositories.ICaretakerRepository, repoDoctor repositories.IDoctorRepository) IAuthService {
	return &authService{
		UsersRepository:     repoUsers,
		OwnerRepository:     repoOwner,
		CaretakerRepository: repoCaretaker,
		DoctorRepository:    repoDoctor,
	}
}

func (sv *authService) CheckToken(td *middlewares.TokenDetails) error {
	if _, err := sv.UsersRepository.FindByID(td.UserID); err != nil {
		return err
	}
	return nil
}

func (sv *authService) Register(role string, data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	var userData *entities.UserDataModel
	var err error
	userData, err = sv.UsersRepository.InsertUser(role, data)
	if err != nil {
		return nil, err
	}
	switch role {
	case "doctor":
		userData, err = sv.DoctorRepository.InsertDoctor(userData)
		if err != nil {
			return nil, err
		}
	case "caretaker":
		userData, err = sv.CaretakerRepository.InsertCaretaker(userData)
		if err != nil {
			return nil, err
		}
	case "owner":
		userData, err = sv.OwnerRepository.InsertOwner(userData)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("role is required")
	}
	return userData, nil
}

func (sv *authService) Login(role string, data entities.LoginUserRequestModel) (*entities.LoginUserResponseModel, error) {
	userData, err := sv.UsersRepository.FindByEmailAndRole(data.Email, role)
	if err != nil {
		return nil, err
	}
	if !utils.CheckPasswordHash(data.Password, userData.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	return userData, nil
}

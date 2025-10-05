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
	var userData, roleData *entities.UserDataModel
	var err error
	userData, err = sv.UsersRepository.InsertUser(role, data)
	if err != nil {
		return nil, err
	}
	switch role {
	case "doctor":
		roleData, err = sv.DoctorRepository.InsertDoctor(userData.UserID, userData.LicenseNumber)
		if err != nil {
			return nil, err
		}
	case "caretaker":
		roleData, err = sv.CaretakerRepository.InsertCaretaker(userData.UserID, userData.Specialization)
		if err != nil {
			return nil, err
		}
	case "owner":
		roleData, err = sv.OwnerRepository.InsertOwner(userData.UserID)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("role is required")
	}
	if role != "admin" && userData.UserID != roleData.UserID {
		return nil, fmt.Errorf("invalid foreign key user_id")
	}
	userData.LicenseNumber = roleData.LicenseNumber
	userData.Specialization = roleData.Specialization
	userData.TotalSpending = roleData.TotalSpending
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

package services

import (
	"fmt"
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"
	"time"
)

type authService struct {
	AdminRepository     repositories.IAdminRepository
	OwnerRepository     repositories.IOwnerRepository
	CaretakerRepository repositories.ICaretakerRepository
	DoctorRepository    repositories.IDoctorRepository
}

type IAuthService interface {
	CheckToken(td *middlewares.TokenDetails) error
	Register(role string, data entities.CreatedUserModel) (*entities.UserDataModel, error)
	Login(role string, data entities.LoginUserModel) (*entities.LoginUserModel, error)
}

func NewAuthService(repoAdmin repositories.IAdminRepository, repoOwner repositories.IOwnerRepository, repoCaretaker repositories.ICaretakerRepository, repoDoctor repositories.IDoctorRepository) IAuthService {
	return &authService{
		AdminRepository:     repoAdmin,
		OwnerRepository:     repoOwner,
		CaretakerRepository: repoCaretaker,
		DoctorRepository:    repoDoctor,
	}
}

func (sv *authService) CheckToken(td *middlewares.TokenDetails) error {
	switch td.Role {
	case "admin":
		if _, err := sv.AdminRepository.FindByID(td.UserID); err != nil {
			return err
		}
	case "doctor":
		if _, err := sv.DoctorRepository.FindByID(td.UserID); err != nil {
			return err
		}
	case "caretaker":
		if _, err := sv.CaretakerRepository.FindByID(td.UserID); err != nil {
			return err
		}
	case "owner":
		if _, err := sv.OwnerRepository.FindByID(td.UserID); err != nil {
			return err
		}
	}
	return nil
}

func (sv *authService) Register(role string, data entities.CreatedUserModel) (*entities.UserDataModel, error) {
	now := time.Now()
	eighteenYearsLater := data.BirthDate.AddDate(18, 0, 0)
	if now.Before(eighteenYearsLater) {
		return nil, fmt.Errorf("you must be at least 18 years old to register")
	}

	if len(data.TelephoneNumber) != 10 {
		return nil, fmt.Errorf("telephone number must be 10 digits")
	}

	var userData *entities.UserDataModel
	var err error
	switch role {
	case "admin":
		userData, err = sv.AdminRepository.InsertAdmin(data)
		if err != nil {
			return nil, err
		}
	case "doctor":
		userData, err = sv.DoctorRepository.InsertDoctor(data)
		if err != nil {
			return nil, err
		}
	case "caretaker":
		userData, err = sv.CaretakerRepository.InsertCaretaker(data)
		if err != nil {
			return nil, err
		}
	case "owner":
		userData, err = sv.OwnerRepository.InsertOwner(data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("role is required")
	}
	return userData, nil
}

func (sv *authService) Login(role string, data entities.LoginUserModel) (*entities.LoginUserModel, error) {
	var userData *entities.LoginUserModel
	var err error
	switch role {
	case "admin":
		userData, err = sv.AdminRepository.FindByEmail(data.Email)
		if err != nil {
			return nil, err
		}
	case "doctor":
		userData, err = sv.DoctorRepository.FindByEmail(data.Email)
		if err != nil {
			return nil, err
		}
	case "caretaker":
		userData, err = sv.CaretakerRepository.FindByEmail(data.Email)
		if err != nil {
			return nil, err
		}
	case "owner":
		userData, err = sv.OwnerRepository.FindByEmail(data.Email)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("role is required")
	}
	if !utils.CheckPasswordHash(data.Password, userData.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	return userData, nil
}

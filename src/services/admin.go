package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/repositories"
)

type AdminService struct {
	Repo repositories.IAdminRepository
}

type IAdminService interface {
	FindAdminByID(id string) (*entities.UserDataModel, error)
	DeleteAdminByID(id string) (*entities.UserDataModel, error)
	UpdateAdminByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error)
}

func NewAdminService(repo repositories.IAdminRepository) IAdminService {
	return &AdminService{Repo: repo}
}

func (s *AdminService) FindAdminByID(id string) (*entities.UserDataModel, error) {
	// You may want to add validation or logging here
	return s.Repo.FindByID(id)
}

func (s *AdminService) DeleteAdminByID(id string) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.DeleteByID(id)
}

func (s *AdminService) UpdateAdminByID(id string, data entities.UpdateUserModel) (*entities.UserDataModel, error) {
	// You could add validation, logging, or pre-checks here
	return s.Repo.UpdateByID(id, data)
}

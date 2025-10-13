package services

import (
	"fmt"

	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/domain/repositories"
)

type ServiceService struct {
	Repo repositories.IServiceRepository
}

type IServiceService interface {
	CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, error)
	DeleteServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServiceByID(serviceID string) (*entities.ServiceModel, error)
}

func NewServiceService(repo repositories.IServiceRepository) IServiceService {
	return &ServiceService{Repo: repo}
}

func (s *ServiceService) CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, error) {
	status := db.ServiceStatus(data.Status)
	switch status {
	case db.ServiceStatusWait, db.ServiceStatusOngoing, db.ServiceStatusFinish:
	default:
		return nil, fmt.Errorf("service -> CreateService: invalid status %q", data.Status)
	}

	return s.Repo.Insert(data)
}

func (s *ServiceService) DeleteServiceByID(serviceID string) (*entities.ServiceModel, error) {
	return s.Repo.DeleteByID(serviceID)
}

func (s *ServiceService) FindServiceByID(serviceID string) (*entities.ServiceModel, error) {
	return s.Repo.FindByID(serviceID)
}

package services

import (
	"fmt"
	"strings"

	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/domain/repositories"
)

type ServiceService struct {
	Repo          repositories.IServiceRepository
	CaretakerRepo repositories.ICaretakerRepository
	DoctorRepo    repositories.IDoctorRepository
}

type IServiceService interface {
	CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, error)
	UpdateServiceByID(serviceID string, data entities.UpdateServiceRequest) (*entities.ServiceModel, error)
	DeleteServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServicesByOwnerID(ownerID string, status string) ([]*entities.ServiceModel, error)
	FindAllServices(status string) ([]*entities.ServiceModel, error)
}

func NewServiceService(
    repo repositories.IServiceRepository,
    caretakerRepo repositories.ICaretakerRepository,
    doctorRepo repositories.IDoctorRepository,
) IServiceService {
    return &ServiceService{
        Repo:          repo,
        CaretakerRepo: caretakerRepo,
        DoctorRepo:    doctorRepo,
    }
}


func (s *ServiceService) CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, error) {
	status := db.ServiceStatus(data.Status)
	switch status {
	case db.ServiceStatusWait, db.ServiceStatusOngoing, db.ServiceStatusFinish:
	default:
		return nil, fmt.Errorf("service -> CreateService: invalid status %q", data.Status)
	}

	serviceType := strings.ToLower(strings.TrimSpace(data.ServiceType))
	switch serviceType {
	case "cservice":
		if _, err := s.CaretakerRepo.FindByID(data.StaffID); err != nil {
			return nil, fmt.Errorf("service -> CreateService: caretaker not found: %w", err)
		}
		if data.Comment != nil {
			trimmed := strings.TrimSpace(*data.Comment)
			if trimmed == "" {
				data.Comment = nil
			} else {
				data.Comment = &trimmed
			}
		}
	case "mservice":
		if _, err := s.DoctorRepo.FindByID(data.StaffID); err != nil {
			return nil, fmt.Errorf("service -> CreateService: doctor not found: %w", err)
		}
		if data.Disease == nil || strings.TrimSpace(*data.Disease) == "" {
			return nil, fmt.Errorf("service -> CreateService: disease is required for mservice")
		}
		trimmed := strings.TrimSpace(*data.Disease)
		data.Disease = &trimmed
	default:
		return nil, fmt.Errorf("service -> CreateService: invalid service_type %q", data.ServiceType)
	}

	data.ServiceType = serviceType

	return s.Repo.Insert(data)
}

func (s *ServiceService) UpdateServiceByID(serviceID string, data entities.UpdateServiceRequest) (*entities.ServiceModel, error) {
	if data.Status != nil {
		status := db.ServiceStatus(*data.Status)
		switch status {
		case db.ServiceStatusWait, db.ServiceStatusOngoing, db.ServiceStatusFinish:
		default:
			return nil, fmt.Errorf("service -> UpdateServiceByID: invalid status %q", *data.Status)
		}
	}

	currentService, err := s.Repo.FindByID(serviceID)
	if err != nil {
		return nil, err
	}

	var targetType string
	if data.ServiceType != nil && strings.TrimSpace(*data.ServiceType) != "" {
		serviceType := strings.ToLower(strings.TrimSpace(*data.ServiceType))
		if serviceType != "cservice" && serviceType != "mservice" {
			return nil, fmt.Errorf("service -> UpdateServiceByID: invalid service_type %q", *data.ServiceType)
		}
		data.ServiceType = &serviceType
		targetType = serviceType
	} else {
		targetType = currentService.ServiceType
	}

	switch targetType {
	case "cservice":
		if data.StaffID != nil {
			if _, err := s.CaretakerRepo.FindByID(*data.StaffID); err != nil {
				return nil, fmt.Errorf("service -> UpdateServiceByID: caretaker not found: %w", err)
			}
		}
		if currentService.ServiceType != "cservice" && data.StaffID == nil {
			return nil, fmt.Errorf("service -> UpdateServiceByID: staff_id required when switching to cservice")
		}
		if data.Comment != nil {
			trimmed := strings.TrimSpace(*data.Comment)
			if trimmed == "" {
				data.Comment = nil
			} else {
				data.Comment = &trimmed
			}
		}

	case "mservice":
		if data.StaffID != nil {
			if _, err := s.DoctorRepo.FindByID(*data.StaffID); err != nil {
				return nil, fmt.Errorf("service -> UpdateServiceByID: doctor not found: %w", err)
			}
		}
		if currentService.ServiceType != "mservice" {
			if data.StaffID == nil {
				return nil, fmt.Errorf("service -> UpdateServiceByID: staff_id required when switching to mservice")
			}
			if data.Disease == nil || strings.TrimSpace(*data.Disease) == "" {
				return nil, fmt.Errorf("service -> UpdateServiceByID: disease required when switching to mservice")
			}
		}
		if data.Disease != nil {
			trimmed := strings.TrimSpace(*data.Disease)
			if trimmed == "" {
				return nil, fmt.Errorf("service -> UpdateServiceByID: disease cannot be empty")
			}
			data.Disease = &trimmed
		}

	case "":
		return nil, fmt.Errorf("service -> UpdateServiceByID: target service type is unknown")
	default:
		return nil, fmt.Errorf("service -> UpdateServiceByID: invalid target service type %q", targetType)
	}

	return s.Repo.UpdateByID(serviceID, data)
}

func (s *ServiceService) DeleteServiceByID(serviceID string) (*entities.ServiceModel, error) {
	return s.Repo.DeleteByID(serviceID)
}

func (s *ServiceService) FindServiceByID(serviceID string) (*entities.ServiceModel, error) {
	return s.Repo.FindByID(serviceID)
}

func (s *ServiceService) FindServicesByOwnerID(ownerID string, status string) ([]*entities.ServiceModel, error) {
    return s.Repo.FindByOwnerID(ownerID, status)
}
func (s *ServiceService) FindAllServices(status string) ([]*entities.ServiceModel, error) {
    return s.Repo.FindAll(status)
}

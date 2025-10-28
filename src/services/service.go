package services

import (
	"fmt"
	"sort"
	"time"

	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/domain/repositories"
	"lama-backend/src/utils"
)

type ServiceService struct {
	Repo          repositories.IServiceRepository
	CaretakerRepo repositories.ICaretakerRepository
	DoctorRepo    repositories.IDoctorRepository
	MserviceRepo  repositories.IMServiceRepository
	CserviceRepo  repositories.ICServiceRepository
	PaymentRepo   repositories.IPaymentRepository
}

type IServiceService interface {
	CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, error)
	UpdateServiceByID(serviceID string, data entities.UpdateServiceRequest) (*entities.ServiceModel, error)
	DeleteServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServicesByOwnerID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	FindServicesByDoctorID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	FindServicesByCaretakerID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	FindAllServices(status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	UpdateStatus(serviceID, status, role, userID string) error
	FindAvailableStaff(serviceType string, date time.Time, page, limit int) ([]*entities.AvailableStaffResponse, int, error)
}

func NewServiceService(
	repo repositories.IServiceRepository,
	caretakerRepo repositories.ICaretakerRepository,
	doctorRepo repositories.IDoctorRepository,
	mserviceRepo repositories.IMServiceRepository,
	cserviceRepo repositories.ICServiceRepository,
	paymentRepo repositories.IPaymentRepository,
) IServiceService {
	return &ServiceService{
		Repo:          repo,
		CaretakerRepo: caretakerRepo,
		DoctorRepo:    doctorRepo,
		MserviceRepo:  mserviceRepo,
		CserviceRepo:  cserviceRepo,
		PaymentRepo:   paymentRepo,
	}
}

func (s *ServiceService) CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, error) {
	status := db.ServiceStatus(data.Status)
	switch status {
	case db.ServiceStatusWait, db.ServiceStatusOngoing, db.ServiceStatusFinish:
	default:
		return nil, fmt.Errorf("service -> CreateService: invalid status %q", data.Status)
	}

	var result *entities.ServiceModel
	switch data.ServiceType {
	case "cservice":
		if _, err := s.CaretakerRepo.FindByID(data.StaffID); err != nil {
			return nil, fmt.Errorf("service -> CreateService: caretaker not found: %w", err)
		}
		payment, err := s.PaymentRepo.InsertPayment(data.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to create payment: %w", err)
		}
		data.PaymentID = payment.PayID
		if result, err = s.Repo.Insert(data); err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to create service: %w", err)
		}
		if _, err = s.CserviceRepo.Insert(*mapToSubService(*result)); err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to create cservice: %w", err)
		}
	case "mservice":
		if _, err := s.DoctorRepo.FindByID(data.StaffID); err != nil {
			return nil, fmt.Errorf("service -> CreateService: doctor not found: %w", err)
		}
		payment, err := s.PaymentRepo.InsertPayment(data.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to create payment: %w", err)
		}
		data.PaymentID = payment.PayID
		if result, err = s.Repo.Insert(data); err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to create service: %w", err)
		}
		if _, err = s.MserviceRepo.Insert(*mapToSubService(*result)); err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to create mservice: %w", err)
		}
	default:
		return nil, fmt.Errorf("service -> CreateService: invalid service_type %q", data.ServiceType)
	}
	return result, nil
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

	var result *entities.ServiceModel
	switch currentService.ServiceType {
	case "cservice":
		if data.StaffID != nil {
			if _, err := s.CaretakerRepo.FindByID(*data.StaffID); err != nil {
				return nil, fmt.Errorf("service -> UpdateServiceByID: caretaker not found: %w", err)
			}
		}
		if result, err = s.Repo.UpdateByID(serviceID, data); err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to update service: %w", err)
		}
		subResult, err := s.CserviceRepo.UpdateByID(*mapToSubService(*result))
		if err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to update cservice: %w", err)
		}
		result.ServiceType = "cservice"
		result.StaffID = subResult.StaffID
	case "mservice":
		if data.StaffID != nil {
			if _, err := s.DoctorRepo.FindByID(*data.StaffID); err != nil {
				return nil, fmt.Errorf("service -> UpdateServiceByID: doctor not found: %w", err)
			}
		}
		if result, err = s.Repo.UpdateByID(serviceID, data); err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to update service: %w", err)
		}
		subResult, err := s.MserviceRepo.UpdateByID(*mapToSubService(*result))
		if err != nil {
			return nil, fmt.Errorf("service -> CreateService: failed to update mservice: %w", err)
		}
		result.ServiceType = "mservice"
		result.StaffID = subResult.StaffID
	default:
		return nil, fmt.Errorf("service -> UpdateServiceByID: invalid target service type")
	}

	return result, nil
}

func (s *ServiceService) DeleteServiceByID(serviceID string) (*entities.ServiceModel, error) {
	return s.Repo.DeleteByID(serviceID)
}

func (s *ServiceService) FindServiceByID(serviceID string) (*entities.ServiceModel, error) {
	return s.Repo.FindByID(serviceID)
}

func (s *ServiceService) FindServicesByOwnerID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.Repo.FindByOwnerID(ownerID, status, month, year, offset, limit)
}

func (s *ServiceService) FindServicesByDoctorID(doctorID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.Repo.FindByDoctorID(doctorID, status, month, year, offset, limit)
}

func (s *ServiceService) FindServicesByCaretakerID(caretakerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.Repo.FindByCaretakerID(caretakerID, status, month, year, offset, limit)
}

func (s *ServiceService) FindAllServices(status string, month, year, page int, limit int) ([]*entities.ServiceModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.Repo.FindAll(status, month, year, offset, limit)
}

func (s *ServiceService) UpdateStatus(serviceID, status, role, userID string) error {
	service, err := s.Repo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("service -> UpdateStatus: %w", err)
	}

	switch role {
	case "admin":
		// Admin can update any service
	case "caretaker":
		if service.ServiceType != "cservice" {
			return fmt.Errorf("service -> UpdateStatus: caretaker can only update cservice")
		}
		if service.StaffID != userID {
			return fmt.Errorf("service -> UpdateStatus: caretaker can only update their own services")
		}
	case "doctor":
		if service.ServiceType != "mservice" {
			return fmt.Errorf("service -> UpdateStatus: doctor can only update mservice")
		}
		if service.StaffID != userID {
			return fmt.Errorf("service -> UpdateStatus: doctor can only update their own services")
		}
	default:
		return fmt.Errorf("service -> UpdateStatus: invalid role %q", role)
	}
	return s.Repo.UpdateStatus(serviceID, status)
}

func (s *ServiceService) FindAvailableStaff(serviceType string, date time.Time, page, limit int) ([]*entities.AvailableStaffResponse, int, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	switch serviceType {
	case "caretaker":
		caretakers, amount, err := s.CaretakerRepo.FindAvailableCaretaker(date, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		var results []*entities.AvailableStaffResponse
		for _, c := range *caretakers {
			cservices := c.Cservice()

			// use a set to collect unique busy hours
			hourSet := map[int]bool{}

			for _, cs := range cservices {
				service := cs.Service()
				if service != nil && utils.CheckSameDate(service.RdateStart, date) {
					startHour := service.RdateStart.Hour()
					endHour := service.RdateEnd.Hour()

					for h := startHour; h < endHour; h++ {
						hourSet[h] = true // mark hour as busy
					}
				}
			}

			if len(hourSet) >= 8 {
				continue
			}

			// convert unique hours to slice
			var busyTimeSlot []int
			for h := range hourSet {
				busyTimeSlot = append(busyTimeSlot, h)
			}
			sort.Ints(busyTimeSlot)

			userData := c.Users()
			rating, _ := c.Rating()
			// you can now use busyTimeSlot here
			results = append(results, &entities.AvailableStaffResponse{
				ID:           c.UserID,
				Name:         userData.Name,
				BusyTimeSlot: busyTimeSlot,
				Rating:       rating,
			})
		}
		return results, amount, err
	case "doctor":
		return nil, 0, nil
	default:
		return nil, 0, nil
	}
}

func mapToSubService(service entities.ServiceModel) *entities.SubService {
	result := &entities.SubService{
		ServiceID: service.Sid,
		StaffID:   service.StaffID,
		Disease:   service.Disease,
		Comment:   service.Comment,
		Score:     service.Score,
	}
	return result
}

func calDefaultLimitAndOffset(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 5
	}
	//return offset, limit
	return (page - 1) * limit, limit
}

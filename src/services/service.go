package services

import (
	"fmt"
	"math"
	"strings"
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
	ValidateServiceCreation(data entities.CreateServiceRequest, payment_status string) error
	CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, *entities.SubService, error)
	UpdateServiceByID(serviceID string, data entities.UpdateServiceRequest) (*entities.ServiceModel, error)
	DeleteServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServiceByID(serviceID string) (*entities.ServiceModel, error)
	FindServicesByOwnerID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	FindServicesByDoctorID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	FindServicesByCaretakerID(ownerID string, status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	FindAllServices(status string, month, year, page int, limit int) ([]*entities.ServiceModel, error)
	UpdateStatus(serviceID, status, role, userID string) error
	FindAvailableStaff(serviceType string, startDate, endDate time.Time) ([]*entities.AvailableStaffResponse, error)
	FindBusyTimeSlot(serviceType string, staffID string, startDate00, startDate23, endDate00, endDate23 time.Time) (map[string][]string, error)
	GetScoreAndReviewByCaretakerID(caretakerID string) (float64, []*entities.SubService, error)
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

func (s *ServiceService) ValidateServiceCreation(data entities.CreateServiceRequest, payment_status string) error {
	// status exist
	status := db.ServiceStatus(data.Status)
	validStatuses := map[db.ServiceStatus]bool{
		db.ServiceStatusWait:    true,
		db.ServiceStatusOngoing: true,
		db.ServiceStatusFinish:  true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("service -> CreateServiceStripe: invalid status %q", data.Status)
	}

	// staff exist
	switch data.ServiceType {
	case "cservice":
		if _, err := s.CaretakerRepo.FindByID(data.StaffID); err != nil {
			return fmt.Errorf("service -> CreateServiceStripe: caretaker not found: %w", err)
		}
	case "mservice":
		if _, err := s.DoctorRepo.FindByID(data.StaffID); err != nil {
			return fmt.Errorf("service -> CreateServiceStripe: doctor not found: %w", err)
		}
	default:
		return fmt.Errorf("service -> CreateServiceStripe: invalid service_type %q", data.ServiceType)
	}

	// payment exist
	payment, err := s.PaymentRepo.FindByID(data.PaymentID)
	if err != nil {
		return fmt.Errorf("service -> CreateServiceStripe: failed to receive payment: %w", err)
	}
	if payment.OwnerID != data.OwnerID {
		return fmt.Errorf("service -> CreateServiceStripe: service owner and payment owner have to be same person")
	}

	// payment status correct
	switch payment_status {
	case "unpaid":
		if payment.Status != db.PaymentStatusUnpaid {
			return fmt.Errorf("service -> CreateServiceStripe: payment must be UnPaid before pay")
		}
	case "paid":
		if payment.Status != db.PaymentStatusPaid {
			return fmt.Errorf("service -> CreateServiceStripe: payment must be Paid before create service")
		}
	default:
		return fmt.Errorf("service -> CreateServiceStripe: invalid payment status %q", payment_status)
	}

	return nil
}

func (s *ServiceService) CreateService(data entities.CreateServiceRequest) (*entities.ServiceModel, *entities.SubService, error) {
	if err := s.ValidateServiceCreation(data, "paid"); err != nil {
		return nil, nil, err
	}

	var service *entities.ServiceModel
	var subService *entities.SubService
	var err error
	switch data.ServiceType {
	case "cservice":
		// insert service
		if service, err = s.Repo.Insert(data); err != nil {
			return nil, nil, fmt.Errorf("service -> CreateService: failed to create service: %w", err)
		}

		// insert subservice
		if subService, err = s.CserviceRepo.Insert(*mapToSubService(*service)); err != nil {
			return nil, nil, fmt.Errorf("service -> CreateService: failed to create cservice: %w", err)
		}
	case "mservice":
		if service, err = s.Repo.Insert(data); err != nil {
			return nil, nil, fmt.Errorf("service -> CreateService: failed to create service: %w", err)
		}

		service.Disease = data.Disease
		if subService, err = s.MserviceRepo.Insert(*mapToSubService(*service)); err != nil {
			return nil, nil, fmt.Errorf("service -> CreateService: failed to create mservice: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("service -> CreateService: invalid service_type %q", data.ServiceType)
	}
	return service, subService, nil
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

func (s *ServiceService) FindAvailableStaff(serviceType string, startDate, endDate time.Time) ([]*entities.AvailableStaffResponse, error) {
	var staff []*entities.AvailableStaffResponse
	var err error
	switch serviceType {
	case "cservice":
		staff, err = s.CaretakerRepo.FindAvailableCaretaker(startDate, endDate)
		if err != nil {
			return nil, err
		}
	case "mservice":
		staff, err = s.DoctorRepo.FindAvailableDoctor(startDate, endDate)
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	return staff, nil
}

func (s *ServiceService) FindBusyTimeSlot(
	serviceType string,
	staffID string,
	startDate00, startDate23, endDate00, endDate23 time.Time,
) (map[string][]string, error) {
	var services *[]db.ServiceModel
	var err error
	switch serviceType {
	case "cservice":
		services, err = s.CaretakerRepo.FindBusyTimeSlot(staffID, startDate00, startDate23, endDate00, endDate23)
	case "mservice":
		services, err = s.DoctorRepo.FindBusyTimeSlot(staffID, startDate00, startDate23, endDate00, endDate23)
	default:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	workingStart := 8
	workingEnd := 17

	result := make(map[string][]string)

	for _, svc := range *services {
		svcStart := svc.RdateStart
		svcEnd := svc.RdateEnd

		// clip service time within request date range
		busyStart := utils.MaxTime(svcStart, startDate00)
		busyEnd := utils.MinTime(svcEnd, endDate23)

		if !busyStart.Before(busyEnd) {
			continue // no overlap
		}

		// iterate each day ONLY within request range
		for day := busyStart; day.Before(busyEnd); day = day.Add(24 * time.Hour) {

			dateKey := day.Format("2006-01-02")

			// skip days that are outside request range
			if day.Before(startDate00) || day.After(endDate23) {
				continue
			}

			// working hour limits
			dayStart := time.Date(day.Year(), day.Month(), day.Day(), workingStart, 0, 0, 0, day.Location())
			dayEnd := time.Date(day.Year(), day.Month(), day.Day(), workingEnd, 0, 0, 0, day.Location())

			realStart := utils.MaxTime(busyStart, dayStart)
			realEnd := utils.MinTime(busyEnd, dayEnd)

			if !realStart.Before(realEnd) {
				continue
			}

			// add hourly busy slots
			for t := realStart; t.Before(realEnd); t = t.Add(time.Hour) {
				result[dateKey] = append(result[dateKey], fmt.Sprintf("%02d:00", t.Hour()))
			}
		}
	}

	return result, nil
}

func (s *ServiceService) GetScoreAndReviewByCaretakerID(caretakerID string) (float64, []*entities.SubService, error) {
	subservices, err := s.CserviceRepo.FindByCaretakerID(caretakerID)
	if err != nil {
		return 0.0, nil, err
	}

	var total int
	var count int
	var reviews []*entities.SubService

	for _, ss := range subservices {
		if ss.Score != nil {
			total += *ss.Score
			count++
		}
		if ss.Comment != nil {
			trimmed := strings.TrimSpace(*ss.Comment)
			if trimmed != "" {
				c := trimmed
				reviews = append(reviews, &entities.SubService{
					ServiceID: ss.ServiceID,
					StaffID:   ss.StaffID,
					Comment:   &c,
					Score:     ss.Score,
				})
			}
		}
	}

	if count == 0 {
		return 0.0, reviews, nil
	}

	avg := float64(total) / float64(count)
	avg = math.Round(avg*10) / 10
	return avg, reviews, nil
}

// helper to compare date only (ignore time)
func sameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
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

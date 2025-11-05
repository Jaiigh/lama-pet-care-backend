package services

import (
	

	"lama-backend/domain/entities"
	
	"lama-backend/domain/repositories"
)

type PaymentService struct {
	repo repositories.IPaymentRepository
}

type IPaymentService interface {
	FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, error)
	FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, error)
}

func NewPaymentService(repo repositories.IPaymentRepository) IPaymentService {
	return &PaymentService{
		repo: repo,
	}
}
func (s *PaymentService) FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.repo.FindAllPayments(month, year, offset, limit)
}

func (s *PaymentService) FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.repo.FindPaymentsByOwnerID(ownerID, month, year, offset, limit)
}


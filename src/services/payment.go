package services

import (
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/domain/repositories"
	"strings"
	"time"
)

type PaymentService struct {
	repo repositories.IPaymentRepository
}

type IPaymentService interface {
	InsertPayment(userID string) (*entities.PaymentModel, error)
	FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, error)
	FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, error)
	UpdateByID(paymentID string, data entities.UpdatePaymentRequest) (*entities.PaymentModel, error)
}

func NewPaymentService(repo repositories.IPaymentRepository) IPaymentService {
	return &PaymentService{
		repo: repo,
	}
}

func (s *PaymentService) InsertPayment(userID string) (*entities.PaymentModel, error) {
	return s.repo.InsertPayment(userID)
}

func (s *PaymentService) FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.repo.FindAllPayments(month, year, offset, limit)
}

func (s *PaymentService) FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, error) {
	offset, limit := calDefaultLimitAndOffset(page, limit)
	return s.repo.FindPaymentsByOwnerID(ownerID, month, year, offset, limit)
}

func (s *PaymentService) UpdateByID(paymentID string, data entities.UpdatePaymentRequest) (*entities.PaymentModel, error) {
	paymentModelToRepo := entities.PaymentModel{}

	// ย้าย Logic การ Parse/Trim มาไว้ที่นี่
	// (เพราะ Handler (Gateway) ไม่ควรทำ Logic นี้)
	if data.Status != nil {
		statusString := strings.TrimSpace(*data.Status) // นี่คือ string (เช่น "PAID")
		if statusString != "" {

			// [!! แก้ไข !!]
			// คุณต้อง Cast string ("PAID") ให้เป็น Type enum
			// (db.PaymentStatus) ก่อนยัดใส่ Model
			paymentModelToRepo.Status = db.PaymentStatus(statusString)
		}
	}
	if data.Type != nil {
		typeStr := strings.TrimSpace(*data.Type)
		if typeStr != "" {
			// ถ้า Type ใน PaymentModel เป็น *string (Pointer)
			paymentModelToRepo.Type = &typeStr

			// แต่ถ้า Type ใน PaymentModel เป็น string (ไม่ใช่ Pointer)
			// paymentModelToRepo.Type = typeStr
		}
	}
	if data.PayDate != nil {
		// Parse ที่นี่ (Validator ใน Handler เช็กให้แล้วว่า Format ถูก)
		t, _ := time.Parse(time.RFC3339, *data.PayDate)
		paymentModelToRepo.PayDate = &t
	}

	// [!! แก้ไข !!]
	// ส่ง Model ที่แปลงแล้ว (paymentModelToRepo) ไปให้ Repo
	return s.repo.UpdateByID(paymentID, paymentModelToRepo)
}

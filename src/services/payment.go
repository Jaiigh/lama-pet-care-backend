package services

import (
	"fmt"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/domain/repositories"
	"lama-backend/src/utils"
	"os"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type PaymentService struct {
	repo repositories.IPaymentRepository
}

type IPaymentService interface {
	InsertPayment(userID string, reserve_date_end, reserve_date_start time.Time) (*entities.PaymentModel, error)
	FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, int, error)
	FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, int, error)
	UpdateByID(paymentID string, data entities.UpdatePaymentRequest) (*entities.PaymentModel, error)
	StripeCreatePrice(service *entities.CreateServiceRequest, price int) (string, error)
	GetMethodAndPaydate(payIntent string) (string, string, error)
}

func NewPaymentService(repo repositories.IPaymentRepository) IPaymentService {
	return &PaymentService{
		repo: repo,
	}
}

func CalPrice(reserve_date_end, reserve_date_start time.Time) int {
	durationHours := reserve_date_end.Sub(reserve_date_start).Hours()
	return int(durationHours * 100)
}

func (s *PaymentService) InsertPayment(userID string, reserve_date_end, reserve_date_start time.Time) (*entities.PaymentModel, error) {
	price := CalPrice(reserve_date_end, reserve_date_start)
	return s.repo.InsertPayment(userID, price)
}

func (s *PaymentService) FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, int, error) {
	payment, err := s.repo.FindAllPayments(month, year)
	if err != nil {
		return nil, 0, fmt.Errorf("payment service -> FindPaymentsByOwnerID: %v", err)
	}

	offset, end := utils.CalDefaultOffsetEnd(page, limit)
	paginate, total := paymentPaginate(payment, offset, end)
	return paginate, total, nil
}

func (s *PaymentService) FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, int, error) {
	payment, err := s.repo.FindPaymentsByOwnerID(ownerID, month, year)
	if err != nil {
		return nil, 0, fmt.Errorf("payment service -> FindPaymentsByOwnerID: %v", err)
	}

	offset, end := utils.CalDefaultOffsetEnd(page, limit)
	paginate, total := paymentPaginate(payment, offset, end)
	return paginate, total, nil
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

// CreateCheckoutSession creates a Stripe Checkout Session
func (s *PaymentService) StripeCreatePrice(service *entities.CreateServiceRequest, price int) (string, error) {
	// prepare data - price, currenct, method (price already in pass)
	currency := "thb"
	paymentMethod := []string{"card", "promptpay"}
	stripe.Key = os.Getenv("STRIPE_KEY")
	url := os.Getenv("STRIPE_REDIRECT")

	var unitPrice int32
	name := fmt.Sprintf("pack %v", price)
	unitPrice = int32(price * 100)
	metaData := map[string]string{
		"owner_id":           service.OwnerID,
		"pet_id":             service.PetID,
		"payment_id":         service.PaymentID,
		"staff_id":           service.StaffID,
		"service_type":       service.ServiceType,
		"status":             string(service.Status),
		"reserve_date_start": service.ReserveDateStart.Format(time.RFC3339),
		"reserve_date_end":   service.ReserveDateEnd.Format(time.RFC3339),
	}

	// Optional disease
	if service.Disease != nil {
		metaData["disease"] = *service.Disease
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice(paymentMethod),

		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(name),
					},
					UnitAmount: stripe.Int64(int64(unitPrice)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:                stripe.String(string(stripe.CheckoutSessionModePayment)),
		ClientReferenceID:   stripe.String(service.OwnerID),
		SuccessURL:          stripe.String(url),
		CancelURL:           stripe.String(os.Getenv("FRONT_REDIRECT_URL_STRIPE")),
		AllowPromotionCodes: stripe.Bool(true),
		ExpiresAt:           stripe.Int64(time.Now().Add(60 * time.Minute).Unix()),
		Metadata:            metaData, // blank - don't have package and salescode
	}
	a, err := session.New(params)
	if err != nil {
		return "", err
	}
	return a.URL, nil
}

func (s *PaymentService) GetMethodAndPaydate(payIntent string) (string, string, error) {
	pi, err := paymentintent.Get(payIntent, nil)
	if err != nil {
		return "", "", err
	}

	payDate := time.Unix(pi.Created, 0).Format(time.RFC3339)

	return pi.PaymentMethodTypes[0], payDate, nil
}

func paymentPaginate(services []*entities.PaymentModel, offset, end int) ([]*entities.PaymentModel, int) {
	total := len(services)
	if offset >= total {
		return []*entities.PaymentModel{}, 0
	}
	if end > total {
		end = total
	}
	paginated := services[offset:end]
	return paginated, total
}

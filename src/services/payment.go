package services

import (
	"fmt"
	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/domain/repositories"
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
	CalPrice(reservedDate *entities.CreatePaymentModel) int
	InsertPayment(userID string, price int) (*entities.PaymentModel, error)
	FindAllPayments(month int, year int, page int, limit int) ([]*entities.PaymentModel, error)
	FindPaymentsByOwnerID(ownerID string, month int, year int, page int, limit int) ([]*entities.PaymentModel, error)
	UpdateByID(paymentID string, data entities.UpdatePaymentRequest) (*entities.PaymentModel, error)
	StripeCreatePrice(userID string, payment *entities.PaymentModel) (string, error)
	GetMethodAndPaydate(payIntent string) (string, string, error)
}

func NewPaymentService(repo repositories.IPaymentRepository) IPaymentService {
	return &PaymentService{
		repo: repo,
	}
}

func (s *PaymentService) CalPrice(reservedDate *entities.CreatePaymentModel) int {
	durationHours := reservedDate.ReserveDateEnd.Sub(reservedDate.ReserveDateStart).Hours()
	return int(durationHours * 100)
}

func (s *PaymentService) InsertPayment(userID string, price int) (*entities.PaymentModel, error) {
	return s.repo.InsertPayment(userID, price)
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

// CreateCheckoutSession creates a Stripe Checkout Session
func (s *PaymentService) StripeCreatePrice(userID string, payment *entities.PaymentModel) (string, error) {
	// prepare data - price, currenct, method
	price := payment.Price
	currency := "thb"
	paymentMethod := []string{"card", "promptpay"}
	stripe.Key = os.Getenv("STRIPE_KEY")
	url := os.Getenv("STRIPE_REDIRECT")

	var unitPrice int32
	name := fmt.Sprintf("pack %v", price)
	unitPrice = int32(price * 100)
	metaData := map[string]string{"user_id": userID, "pay_id": payment.PayID}

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
		ClientReferenceID:   stripe.String(userID),
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

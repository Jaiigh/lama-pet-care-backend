package gateways

import (
	"encoding/json"
	"errors"
	"fmt"
	"lama-backend/domain/entities"
	"time"

	"lama-backend/domain/prisma/db"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v76"
)

// for stripe only
func (h *HTTPGateway) StripeWebhookService(ctx *fiber.Ctx) error {
	payload := ctx.Body()
	event := stripe.Event{}
	err := json.Unmarshal(payload, &event)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Webhook."})
	}

	// check stripe payment_status
	if stripeStatus, ok := event.Data.Object["payment_status"]; !ok || stripeStatus.(string) != "paid" {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "stripe payment_status is not paid"})
	}
	status := "PAID"

	// get method and paydate from payment_intent
	payIntent, ok := event.Data.Object["payment_intent"]
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot get payment_intent from stripe"})
	}
	method, paydate, err := h.PaymentService.GetMethodAndPaydate(payIntent.(string))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot get method or paydate from payment_intent:" + err.Error()})
	}

	// get metadata
	metadataRaw, ok := event.Data.Object["metadata"]
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot get metadata from stripe"})
	}
	metadata, ok := metadataRaw.(map[string]interface{})
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "metadata type assertion failed",
		})
	}

	// get user_id and payment_id from metadata
	payId, ok := metadata["payment_id"]
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "metadata not have pay_id"})
	}

	// update payment
	updateData := entities.UpdatePaymentRequest{
		Status:  &status,
		Type:    &method,
		PayDate: &paydate,
	}

	updatedPayment, err := h.PaymentService.UpdateByID(payId.(string), updateData)
	if err != nil {

		if errors.Is(err, db.ErrNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "payment not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	start, err := time.Parse(time.RFC3339, metadata["reserve_date_start"].(string))
	if err != nil {
		return fmt.Errorf("invalid reserve_date_start: %w", err)
	}

	end, err := time.Parse(time.RFC3339, metadata["reserve_date_end"].(string))
	if err != nil {
		return fmt.Errorf("invalid reserve_date_end: %w", err)
	}
	createService := entities.CreateServiceRequest{
		OwnerID:          metadata["owner_id"].(string),
		PetID:            metadata["pet_id"].(string),
		PaymentID:        updatedPayment.PayID,
		StaffID:          metadata["staff_id"].(string),
		ServiceType:      metadata["service_type"].(string),
		Status:           metadata["status"].(string),
		ReserveDateStart: start,
		ReserveDateEnd:   end,
	}

	service, subservice, err := h.ServiceService.CreateService(createService)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "cannot create service: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "payment updated successfully",
		Data: fiber.Map{
			"payment":    updatedPayment,
			"service":    service,
			"subservice": subservice,
		},
		Status: fiber.StatusOK,
	})
}

package gateways

import (
	"encoding/json"
	"errors"
	"lama-backend/domain/entities"

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
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseModel{Message: "Unauthorization Webhook."})
	}

	// check stripe payment_status
	if stripeStatus, ok := event.Data.Object["payment_status"]; !ok || stripeStatus.(string) != "paid" {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "stripe payment_status is not paid"})
	}
	status := "PAID"

	// get method and paydate from payment_intent
	payIntent, ok := event.Data.Object["payment_intent"]
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot get payment_intent from stripe"})
	}
	method, paydate, err := h.PaymentService.GetMethodAndPaydate(payIntent.(string))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot get method from payment_intent"})
	}

	// get metadata
	metadataRaw, ok := event.Data.Object["metadata"]
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot get metadata from stripe"})
	}
	metadata, ok := metadataRaw.(map[string]interface{})
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{
			Message: "metadata type assertion failed",
		})
	}

	// get user_id and pay_id from metadata
	payId, ok := metadata["pay_id"]
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "metadata not have pay_id"})
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

	// service, err := h.ServiceService.CreateService(req)
	// if err != nil {
	// 	return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
	// 		Message: "cannot create service: " + err.Error(),
	// 	})
	// }

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "payment updated successfully",
		Data:    updatedPayment,
		Status:  fiber.StatusOK,
	})
}

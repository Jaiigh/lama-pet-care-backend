package gateways

import (
	"errors"
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"

	"lama-backend/domain/prisma/db"
	"lama-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary      Get payments
// @Description  Get all payments. Admins can see all payments, while owners can only see their own. Can be filtered by month and year. If only the 'year' query is provided, 'month' will default to 1 (January). Supports pagination.
// @Tags         payment
// @Produce      json
// @Security     BearerAuth
// @Param        month  query int    false "Filter payments by month (1-12)"
// @Param        year   query int    false "Filter payments by year (e.g. 2025)"
// @Param        page   query int    false "Page number for pagination" default(1)
// @Param        limit  query int    false "Number of items per page" default(5)
// @Success      200 {object} entities.ResponseModel "Successfully retrieved payments"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /payments [get]
func (h *HTTPGateway) GetMyPayment(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	month := ctx.QueryInt("month")
	year := ctx.QueryInt("year")
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 5)
	var payments []*entities.PaymentModel

	switch token.Role {
	case "admin":
		payments, err = h.PaymentService.FindAllPayments(month, year, page, limit)
	case "owner":
		payments, err = h.PaymentService.FindPaymentsByOwnerID(token.UserID, month, year, page, limit)

	default:
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token. have to be admin or owner"})
	}
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data: fiber.Map{
			"page":     page,
			"amount":   len(payments),
			"payments": payments,
		},
		Status: fiber.StatusOK,
	})
}

// @Summary      Create payments
// @Description  Create payments. Only user and admin can create payment
// @Tags         payment
// @Produce      json
// @Security     BearerAuth
// @Param        body body entities.CreatePaymentModel true "payment payload include time"
// @Success      200 {object} entities.ResponseModel "Successfully retrieved payments"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /payments [post]
func (h *HTTPGateway) CreatePayment(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	var bodydata entities.CreatePaymentModel
	if err := ctx.BodyParser(&bodydata); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if err := validator.New().Struct(bodydata); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	if bodydata.ReserveDateEnd.Before(bodydata.ReserveDateStart) {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "Reservation end date cannot be earlier than the start date."})
	}

	var payment *entities.PaymentModel

	switch token.Role {
	case "owner", "admin":
		payment, err = h.PaymentService.InsertPayment(token.UserID, &bodydata)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
		}
	default:
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Invalid Role. have to be admin or owner"})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    payment,
		Status:  fiber.StatusOK,
	})
}

// @Summary      Update payment status
// @Description  Update the status of a payment by its ID. Only admins are authorized to perform this action can update only status type and paydate.
// @Tags         payment
// @Produce      json
// @Security     BearerAuth
// @Param        paymentID  path  string  true  "Payment ID"
// @Param        status     path  string  true  "New status for the payment" Enums(paid, unpaid, pending)
// @Success      200 {object} entities.ResponseModel "Successfully updated payment status"
// @Failure      400 {object} entities.ResponseMessage "Bad request"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      403 {object} entities.ResponseMessage "Invalid role"
// @Failure      404 {object} entities.ResponseMessage "Payment not found."
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /payments/{paymentID} [patch]
func (h *HTTPGateway) UpdatePaymentByID(ctx *fiber.Ctx) error {

	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	if token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role (Admin only)"})
	}

	paymentID := ctx.Params("paymentID")
	if paymentID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid payment ID"})
	}

	var updateData entities.UpdatePaymentRequest
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid request body"})
	}

	if updateData.Status == nil && updateData.Type == nil && updateData.PayDate == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "no fields to update"})
	}

	if err := h.Validator.Struct(updateData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	updatedPayment, err := h.PaymentService.UpdateByID(paymentID, updateData)
	if err != nil {

		if errors.Is(err, db.ErrNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "payment not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "payment updated successfully",
		Data:    updatedPayment,
		Status:  fiber.StatusOK,
	})

}

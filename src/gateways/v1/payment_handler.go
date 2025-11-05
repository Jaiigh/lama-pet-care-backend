package gateways

import (
	

	"lama-backend/domain/entities"
	
	"lama-backend/src/middlewares"


	
	"github.com/gofiber/fiber/v2"
)

// @Summary      Get payments
// @Description  Get all payments. Admins can see all payments, while owners can only see their own. Can be filtered by month and year. Supports pagination.
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
		payments, err = h.PaymentService.FindAllPayments( month, year, page, limit)
	case "owner":
		payments, err = h.PaymentService.FindPaymentsByOwnerID(token.UserID,  month, year, page, limit)

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

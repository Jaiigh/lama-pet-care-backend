package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"time"

	"github.com/gofiber/fiber/v2"
)

// @Summary Create Leaveday
// @Description Create Staff leaveday by token and day params (format: YYYY-MM-DD)
// @Tags Leaveday
// @Produce json
// @Param day path string true "leaveday"
// @Success 200 {object} entities.ResponseModel "request successfully"
// @Failure 400 {object} entities.ResponseMessage "Invalid request"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 404 {object} entities.ResponseMessage "User not found."
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /leaveday/{day} [post]
// @Security BearerAuth
func (h *HTTPGateway) CreateLeaveday(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "caretaker" && token.Role != "doctor" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	leavedayStr := ctx.Params("day")
	if leavedayStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid leaveday"})
	}

	leaveday, err := time.Parse("2006-01-02", leavedayStr) // <-- for YYYY-MM-DD format
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid date format, expected YYYY-MM-DD",
		})
	}

	var leavedayData *entities.LeavedayModel
	if leavedayData, err = h.LeavedayService.InsertLeaveday(*token.Token, token.Role, leaveday); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "invalid date format, expected YYYY-MM-DD",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "request successfully",
		Data:    &leavedayData,
		Status:  fiber.StatusOK,
	})
}

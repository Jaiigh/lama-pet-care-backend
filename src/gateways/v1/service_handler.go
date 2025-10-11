package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary create service booking
// @Description owner creates a service booking for a pet
// @Tags service
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Param body body entities.CreateServiceRequest true "service data"
// @Success 201 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /services/ [post]
// @Security BearerAuth
func (h *HTTPGateway) CreateService(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	var req entities.CreateServiceRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	req.OwnerID = token.UserID

	if err := validator.New().Struct(req); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: utils.FormatValidationError(err),
		})
	}

	service, err := h.ServiceService.CreateService(req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "cannot create service: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(entities.ResponseModel{
		Message: "service created",
		Data:    service,
		Status:  fiber.StatusCreated,
	})
}

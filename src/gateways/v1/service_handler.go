package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary create service booking
// @Description owner creates their own booking; admins may create on behalf of an owner by providing owner_id in the payload
// @Tags service
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Param body body entities.CreateServiceRequest true "service data (admins must include owner_id)"
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
	if token.Role != "owner" && token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	var req entities.CreateServiceRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	// default to token.UserID, but admins must supply the owner to assign
	switch token.Role {
	case "owner":
		req.OwnerID = token.UserID
	case "admin":
		if req.OwnerID == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "owner_id is required for admin"})
		}
	}

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

// @Summary delete service booking
// @Description owner deletes their service booking by ID
// @Tags service
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Param id path string true "Service ID"
// @Success 200 {object} entities.ResponseMessage "Delete successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid service ID"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 404 {object} entities.ResponseMessage "Service not found"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /services/{id} [delete]
// @Security BearerAuth
func (h *HTTPGateway) DeleteService(ctx *fiber.Ctx) error {
	// Decode JWT token
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized token",
		})
	}

	// Only owner or admin can proceed
	if token.Role != "owner" && token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{
			Message: "Invalid role",
		})
	}

	// Get service ID from path
	serviceID := ctx.Params("serviceID")
	if serviceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid service ID",
		})
	}

	// Fetch the service first to check ownership
	service, err := h.ServiceService.FindServiceByID(serviceID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{
			Message: "service not found",
		})
	}

	// Check ownership for owner role
	if token.Role == "owner" && service.OwnerID != token.UserID {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{
			Message: "You do not own this service",
		})
	}

	// Delete the service
	deletedService, err := h.ServiceService.DeleteServiceByID(serviceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "cannot delete service: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "service deleted successfully",
		Data:    deletedService,
		Status:  fiber.StatusOK,
	})
}

// @Summary get all services of current user
// @Description get all services for the authenticated user (owner)
// @Tags service
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /services/ [get]
// @Security BearerAuth
func (h *HTTPGateway) GetMyServices(ctx *fiber.Ctx) error {
    token, err := middlewares.DecodeJWTToken(ctx)
    if err != nil {
        return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
    }

    var services []*entities.ServiceModel
    if token.Role == "admin" {
        services, err = h.ServiceService.FindAllServices()
    } else {
        services, err = h.ServiceService.FindServicesByOwnerID(token.UserID)
    }
	

    if err != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
    }
    return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
        Message: "success",
        Data:    services,
        Status:  fiber.StatusOK,
    })
}
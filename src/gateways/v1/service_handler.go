package gateways

import (
	"errors"
	"strings"

	"lama-backend/domain/entities"
	"lama-backend/domain/prisma/db"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary Create caretaker/medical service
// @Description Owners create their own bookings; admins may create on behalf of an owner by providing owner_id. Use service_type=cservice (caretaker) or mservice (doctor) and supply staff_id plus type-specific fields.
// @Tags service
// @Accept json
// @Produce json
// @Param body body entities.CreateServiceRequest true "service payload (admins must include owner_id; mservice requires disease)"
// @Success 201 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /services [post]
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

	req.ServiceType = strings.ToLower(strings.TrimSpace(req.ServiceType))
	if req.StaffID == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "staff_id is required",
		})
	}

	switch req.ServiceType {
	case "cservice":
		if req.Comment != nil {
			trimmed := strings.TrimSpace(*req.Comment)
			if trimmed == "" {
				req.Comment = nil
			} else {
				req.Comment = &trimmed
			}
		}
	case "mservice":
		if req.Disease == nil || strings.TrimSpace(*req.Disease) == "" {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
				Message: "disease is required for mservice",
			})
		}
		trimmed := strings.TrimSpace(*req.Disease)
		req.Disease = &trimmed
	default:
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "service_type must be mservice or cservice",
		})
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

// @Summary Update service booking
// @Description Admin-only endpoint for adjusting service data. Provide the fields that need to change. When switching to `cservice`, include `staff_id` for the caretaker and optionally `comment`. When switching to `mservice`, include doctor `staff_id` และ `disease`.
// @Tags service
// @Accept json
// @Produce json
// @Param serviceID path string true "Service ID"
// @Param body body entities.UpdateServiceRequest true "service update payload"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid request"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 404 {object} entities.ResponseMessage "Service not found"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /services/{id} [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdateService(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	serviceID := ctx.Params("serviceID")
	if serviceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid service ID"})
	}

	var req entities.UpdateServiceRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if req.ServiceType != nil {
		normalized := strings.ToLower(strings.TrimSpace(*req.ServiceType))
		if normalized == "" {
			req.ServiceType = nil
		} else {
			*req.ServiceType = normalized
		}
	}
	if req.Comment != nil {
		trimmed := strings.TrimSpace(*req.Comment)
		if trimmed == "" {
			req.Comment = nil
		} else {
			*req.Comment = trimmed
		}
	}
	if req.Disease != nil {
		trimmed := strings.TrimSpace(*req.Disease)
		if trimmed == "" {
			req.Disease = nil
		} else {
			*req.Disease = trimmed
		}
	}

	if req.OwnerID == nil &&
		req.PetID == nil &&
		req.PaymentID == nil &&
		req.Price == nil &&
		req.Status == nil &&
		req.ReserveDate == nil &&
		req.ServiceType == nil &&
		req.StaffID == nil &&
		req.Disease == nil &&
		req.Comment == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "no fields to update"})
	}

	if err := validator.New().Struct(req); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: utils.FormatValidationError(err),
		})
	}

	updatedService, err := h.ServiceService.UpdateServiceByID(serviceID, req)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNotFound):
			return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "service not found"})
		case strings.Contains(strings.ToLower(err.Error()), "invalid"),
			strings.Contains(strings.ToLower(err.Error()), "required"):
			return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
		default:
			return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "service updated",
		Data:    updatedService,
		Status:  fiber.StatusOK,
	})
}

// @Summary delete service booking
// @Description owner deletes their service booking by ID
// @Tags service
// @Accept json
// @Produce json
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

// @Summary      Get services
// @Description  Get all services for the authenticated user. Admins can see all services. Can be filtered by status.
// @Tags         service
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization header string true "Bearer <JWT token>"
// @Param        status query string false "Filter services by status (e.g. all, wait, ongoing, finish)"
// @Success      200 {object} entities.ResponseModel "Request successful"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /services [get]
func (h *HTTPGateway) GetMyServices(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	statusFilter := ctx.Query("status")
	var services []*entities.ServiceModel
	if token.Role == "admin" {

		services, err = h.ServiceService.FindAllServices(statusFilter)
	} else {

		services, err = h.ServiceService.FindServicesByOwnerID(token.UserID, statusFilter)
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

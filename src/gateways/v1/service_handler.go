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
	if err != nil || token.Purpose != "access" {
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
	if req.ReserveDateEnd.Before(req.ReserveDateStart) {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "Reservation end date cannot be earlier than the start date."})
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
// @Description Admin-only endpoint for adjusting service data. Provide the fields that need to change.
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
// @Router /services/{serviceID} [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdateService(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
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
		req.Price == nil &&
		req.Status == nil &&
		req.ReserveDateStart == nil &&
		req.ReserveDateEnd == nil &&
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
// @Router /services/{serviceID} [delete]
// @Security BearerAuth
func (h *HTTPGateway) DeleteService(ctx *fiber.Ctx) error {
	// Decode JWT token
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
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
// @Description  Get all services for the authenticated user. Admins can see all services. Can be filtered by status, month, and year.
// @Tags         service
// @Produce      json
// @Security     BearerAuth
// @Param        status query string false "Filter services by status (e.g. all, wait, ongoing, finish)"
// @Param        month  query int    false "Filter services by month (1-12)"
// @Param        year   query int    false "Filter services by year (e.g. 2025)"
// @Param        page  query int    false "Page number for pagination" [optional default: 1]
// @Param        limit query int    false "Number of items per page" [optional default: 5]
// @Success      200 {object} entities.ResponseModel "Request successful"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /services [get]
func (h *HTTPGateway) GetMyServices(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	statusFilter := ctx.Query("status")
	month := ctx.QueryInt("month")
	year := ctx.QueryInt("year")
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 5)
	var services []*entities.ServiceModel

	switch token.Role {
	case "admin":
		services, err = h.ServiceService.FindAllServices(statusFilter, month, year, page, limit)
	case "owner":
		services, err = h.ServiceService.FindServicesByOwnerID(token.UserID, statusFilter, month, year, page, limit)
	case "doctor":
		services, err = h.ServiceService.FindServicesByDoctorID(token.UserID, statusFilter, month, year, page, limit)
	case "caretaker":
		services, err = h.ServiceService.FindServicesByCaretakerID(token.UserID, statusFilter, month, year, page, limit)
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data: fiber.Map{
			"page":     page,
			"amount":   len(services),
			"services": services,
		},
		Status: fiber.StatusOK,
	})
}

// @Summary Update service status
// @Description Update the status of a service booking. Allowed roles: admin, caretaker, doctor.
// @Tags service
// @Produce json
// @Param serviceID path string true "Service ID"
// @Param status path string true "service status (wait, ongoing, finish)" Enums(wait, ongoing, finish)
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid request"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 404 {object} entities.ResponseMessage "Service not found"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /services/{serviceID}/{status} [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdateStatusService(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "admin" && token.Role != "caretaker" && token.Role != "doctor" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	serviceID := ctx.Params("serviceID")
	if serviceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid service ID"})
	}
	status := ctx.Params("status")
	if status != "wait" && status != "ongoing" && status != "finish" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid status"})
	}

	err = h.ServiceService.UpdateStatus(serviceID, status, token.Role, token.UserID)
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
		Message: "status updated successfully",
		Data: fiber.Map{
			"service_id": serviceID,
			"status":     status,
		},
		Status: fiber.StatusOK,
	})
}

// @Summary      Get available staff
// @Description  Retrieve all staff members available for a specific service type on a given day.
// @Tags         service
// @Produce      json
// @Security     BearerAuth
// @Param        serviceType   query string true   "Service type to check availability for (cservice or mservice)"
// @Param        startDate     query string true   "service start date (format: YYYY-MM-DD)"
// @Param        endDate       query string true   "service end date (format: YYYY-MM-DD)"
// @Success      200 {object} entities.ResponseModel "Request successful"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      403 {object} entities.ResponseMessage "Invalid role"
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /services/staff [get]
func (h *HTTPGateway) GetAvailableStaff(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "owner" && token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{
			Message: "Invalid role",
		})
	}

	serviceType := ctx.Query("serviceType")
	startDateStr := ctx.Query("startDate")
	endDateStr := ctx.Query("endDate")

	if serviceType != "caretaker" && serviceType != "doctor" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid service type, expected 'caretaker' or 'doctor'",
		})
	}

	startDate, endDate, err := utils.GetRDateRange(startDateStr, endDateStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid date or date format, expected YYYY-MM-DD",
		})
	}

	res, err := h.ServiceService.FindAvailableStaff(serviceType, startDate, endDate)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data: fiber.Map{
			"amount": len(res),
			"staff":  res,
		},
		Status: fiber.StatusOK,
	})
}

// @Summary      Get busy time slot
// @Description  Retrieve all busy time slot for a specific staff on a given day.
// @Tags         service
// @Produce      json
// @Security     BearerAuth
// @Param        serviceType   query string true   "Service type to check availability for (cservice or mservice)"
// @Param        startDate     query string true   "service start date (format: YYYY-MM-DD)"
// @Param        endDate       query string true   "service end date (format: YYYY-MM-DD)"
// @Param        staffID       path  string true   "StaffID"
// @Success      200 {object} entities.ResponseModel  "Request successful"
// @Failure      400 {object} entities.ResponseMessage "Invalid request"
// @Failure      401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure      403 {object} entities.ResponseMessage "Invalid role"
// @Failure      500 {object} entities.ResponseMessage "Internal server error"
// @Router       /services/staff/{staffID}/time [get]
func (h *HTTPGateway) GetBusyTimeSlot(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "owner" && token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{
			Message: "Invalid role",
		})
	}

	serviceType := ctx.Query("serviceType")
	startDateStr := ctx.Query("startDate")
	endDateStr := ctx.Query("endDate")
	staffID := ctx.Params("staffID")

	if serviceType != "caretaker" && serviceType != "doctor" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid service type, expected 'caretaker' or 'doctor'",
		})
	}

	startDate00, startDate23, err := utils.GetRDateRange(startDateStr, startDateStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid date or date format, expected YYYY-MM-DD",
		})
	}
	endDate00, endDate23, err := utils.GetRDateRange(endDateStr, endDateStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "invalid date or date format, expected YYYY-MM-DD",
		})
	}

	res, err := h.ServiceService.FindBusyTimeSlot(serviceType, staffID, startDate00, startDate23, endDate00, endDate23)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    res,
		Status:  fiber.StatusOK,
	})
}

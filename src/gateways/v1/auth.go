package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary Check JWT token validity
// @Description Validates the JWT token passed in the Authorization header
// @Tags Auth
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Success 200 {object} entities.ResponseModel "Token is valid"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 404 {object} entities.ResponseMessage "User not found."
// @Router /auth/check_token [get]
// @Security BearerAuth
func (h *HTTPGateway) checkToken(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if err := h.AuthService.CheckToken(token); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "User not found."})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "Token is valid",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

// @Summary Register
// @Description Register new user except admin
// @Tags Auth
// @Accept json
// @Produce json
// @Param role path string true "Role of the user except admin"
// @Param body body entities.CreatedUserModel true "User data"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /auth/register/:role [post]
func (h *HTTPGateway) Register(ctx *fiber.Ctx) error {
	role := ctx.Params("role")
	if role != "doctor" && role != "caretaker" && role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	bodyData := entities.CreatedUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validator.New().Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}
	eighteenYearsLater := bodyData.BirthDate.AddDate(18, 0, 0)
	if time.Now().Before(eighteenYearsLater) {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "you must be at least 18 years old to register"})
	}
	if role == "doctor" && bodyData.LicenseNumber == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "license_number is required for doctor"})
	}

	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}
	bodyData.Password = hashPassword

	userData, err := h.AuthService.Register(role, bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

// @Summary Login
// @Description Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param role path string true "Role of the user"
// @Param body body entities.LoginUserRequestModel true "email and password"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 401 {object} entities.ResponseMessage "Cannot login user: invalid password or email"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /auth/login/:role [post]
func (h *HTTPGateway) Login(ctx *fiber.Ctx) error {
	role := ctx.Params("role")
	if role != "admin" && role != "doctor" && role != "caretaker" && role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	bodyData := entities.LoginUserRequestModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validator.New().Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	userData, err := h.AuthService.Login(role, bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "cannot login user: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

// @Summary create admin
// @Description create new admin user
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body entities.CreatedUserModel true "Admin user data"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /auth/create_admin [post]
func (h *HTTPGateway) CreateAdmin(ctx *fiber.Ctx) error {
	bodyData := entities.CreatedUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validator.New().Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}
	eighteenYearsLater := bodyData.BirthDate.AddDate(18, 0, 0)
	if time.Now().Before(eighteenYearsLater) {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "you must be at least 18 years old to register"})
	}

	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}
	bodyData.Password = hashPassword

	userData, err := h.AuthService.Register("admin", bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, "admin")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"
	"os"

	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary Check JWT token validity
// @Description Validates the JWT token passed in the Authorization header and decodes it.
// @Tags Auth
// @Produce json
// @Success 200 {object} entities.ResponseModel "Token is valid"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 404 {object} entities.ResponseMessage "User not found."
// @Router /auth/token [get]
// @Security BearerAuth
func (h *HTTPGateway) checkToken(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
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
// @Router /auth/register/{role} [post]
// @Security
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

	if check := utils.ValidPassword(bodyData.Password); !check {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "password must be at least 8 characters long"})
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

	token, err := middlewares.GenerateJWTToken(userData.UserID, role, "access")
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
// @Router /auth/login/{role} [post]
// @Security
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

	token, err := middlewares.GenerateJWTToken(userData.UserID, role, "access")
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
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /auth/admin [post]
// @Security BearerAuth
func (h *HTTPGateway) CreateAdmin(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "admin" {
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

	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}
	bodyData.Password = hashPassword

	userData, err := h.AuthService.Register("admin", bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    userData,
		Status:  fiber.StatusOK,
	})
}

// @Summary forgot password
// @Description forgot password and send reset link to email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body entities.SendEmailModel true "user email and role"
// @Success 200 {object} entities.ResponseMessage "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /auth/password/email [post]
// @Security
func (h *HTTPGateway) ForgotPassword(ctx *fiber.Ctx) error {
	bodyData := entities.SendEmailModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validator.New().Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	userID, err := h.AuthService.ValidateEmailAndRole(&bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot send reset password mail: " + err.Error()})
	}

	token, err := middlewares.GenerateResetPasswordJWTToken(userID, string(bodyData.Role), "reset_password")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "Failed to generate token",
		})
	}

	resetLink := os.Getenv("FORGET_PASSWORD_LINK")
	if err := utils.SendResetEmail(bodyData.Email, resetLink+*token.Token); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send email: " + err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "success",
	})
}

// @Summary reset password
// @Description reset password with token from email
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "token from email"
// @Param body body entities.PasswordModel true "user new password"
// @Success 200 {object} entities.ResponseMessage "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /auth/password [post]
// @Security BearerAuth
func (h *HTTPGateway) ResetPassword(ctx *fiber.Ctx) error {
	emailToken := ctx.Query("token")
	token, err := middlewares.DecodeResetPasswordJWTToken(emailToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	bodyData := entities.PasswordModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validator.New().Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	if check := utils.ValidPassword(bodyData.Password); !check {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "password must be at least 8 characters long"})
	}
	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}

	updatedData := entities.UpdateUserModel{}
	updatedData.Password = &hashPassword
	if _, err := h.UsersService.UpdateUsersByID(token.UserID, updatedData); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot update password: " + err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "success",
	})
}

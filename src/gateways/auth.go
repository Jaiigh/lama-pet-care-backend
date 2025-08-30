package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) checkToken(ctx *fiber.Ctx) error {
	td, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if err := h.AuthService.CheckToken(td); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "User not found."})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "Token is valid",
	})
}

func (h *HTTPGateway) Register(ctx *fiber.Ctx) error {
	role := ctx.Query("role")
	if role != "doctor" && role != "caretaker" && role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	var validate = validator.New()
	bodyData := entities.CreatedUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validate.Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body: validation failed"})
	}
	if role == "doctor" && bodyData.LicenseNumber == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "license_number is required for doctor"})
	}

	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}
	bodyData.Password = hashPassword

	userData, err := h.AuthService.Register(role, bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

func (h *HTTPGateway) Login(ctx *fiber.Ctx) error {
	role := ctx.Query("role")
	if role != "admin" && role != "doctor" && role != "caretaker" && role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	bodyData := entities.LoginUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.Email == "" || bodyData.Password == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	userData, err := h.AuthService.Login(role, bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot login user: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, role)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

func (h *HTTPGateway) CreateAdmin(ctx *fiber.Ctx) error {
	var validate = validator.New()
	bodyData := entities.CreatedUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if err := validate.Struct(bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body: validation failed"})
	}

	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}
	bodyData.Password = hashPassword

	userData, err := h.AuthService.Register("admin", bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, "admin")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate token",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    token,
		Status:  fiber.StatusOK,
	})
}

package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) checkToken(ctx *fiber.Ctx) error {
	td, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	if _, err := h.UserService.GetByID(td.UserID); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "User not found."})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "Token is valid",
	})
}

func (h *HTTPGateway) Register(ctx *fiber.Ctx) error {
	bodyData := entities.CreatedUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.Email == "" || bodyData.Password == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.Role == "" {
		bodyData.Role = "user" // default role
	}

	hashPassword, err := utils.HashPassword(bodyData.Password)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot hash password: " + err.Error()})
	}
	bodyData.Password = hashPassword

	userData, err := h.UserService.InsertNewUser(bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, userData.Role)
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
	bodyData := entities.CreatedUserModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.Email == "" || bodyData.Password == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	userData, err := h.UserService.Login(bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot login user: " + err.Error()})
	}

	token, err := middlewares.GenerateJWTToken(userData.UserID, userData.Role)
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

func (h *HTTPGateway) Logout(ctx *fiber.Ctx) error {
	_, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "logout success",
	})
}

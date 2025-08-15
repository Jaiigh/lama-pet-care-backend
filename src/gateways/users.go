package gateways

import (
	"lama-backend/domain/entities"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) CreateUser(ctx *fiber.Ctx) error {
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

	if _, err := h.UserService.InsertNewUser(bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot insert new user account: " + err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{Message: "success"})
}

func (h *HTTPGateway) GetAllUserData(ctx *fiber.Ctx) error {
	data, err := h.UserService.GetAllUsers()
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot get all users data"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data, Status: 200})
}

func (h *HTTPGateway) GetByID(ctx *fiber.Ctx) error {
	bodyData := entities.UserIDModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	data, err := h.UserService.GetByID(bodyData.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot get user data"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data, Status: 200})
}

func (h *HTTPGateway) UpdateUser(ctx *fiber.Ctx) error {
	bodyData := entities.UserDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.UserID == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "user_id is required"})
	}

	updatedData, err := h.UserService.UpdateUser(bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "cannot update user data: " + err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: updatedData, Status: 200})
}

func (h *HTTPGateway) DeleteUser(ctx *fiber.Ctx) error {
	bodyData := entities.UserIDModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.UserID == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "user_id is required"})
	}

	if err := h.UserService.DeleteUser(bodyData.UserID); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot delete user data: " + err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{Message: "success"})
}

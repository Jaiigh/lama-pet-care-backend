package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) FindOwnerByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	user, err := h.OwnerService.FindOwnerByID(token.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "user not found",
			"detail": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

func (h *HTTPGateway) DeleteOwnerByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	deletedUser, err := h.OwnerService.DeleteOwnerByID(token.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "owner deleted successfully",
		"data":    deletedUser,
	})
}

func (h *HTTPGateway) UpdateOwnerByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	var updateData entities.UpdateUserModel
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updatedUser, err := h.OwnerService.UpdateOwnerByID(token.UserID, updateData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedUser)
}

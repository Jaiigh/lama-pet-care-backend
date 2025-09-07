package gateways

import (
	"github.com/gofiber/fiber/v2"
	"lama-backend/domain/entities"
)

func (h *HTTPGateway) FindDoctorByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	user, err := h.DoctorService.FindDoctorByID(id)
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

func (h *HTTPGateway) DeleteDoctorByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing doctor ID",
		})
	}

	deletedUser, err := h.DoctorService.DeleteDoctorByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "doctor deleted successfully",
		"data":    deletedUser,
	})
}

func (h *HTTPGateway) UpdateDoctorByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing id parameter",
		})
	}

	var updateData entities.UpdateUserModel
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updatedUser, err := h.DoctorService.UpdateDoctorByID(id, updateData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedUser)
}

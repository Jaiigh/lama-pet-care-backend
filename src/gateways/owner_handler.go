package gateways

import (
	"github.com/gofiber/fiber/v2"
)

// FindOwnerByID handles GET /owner/:id requests
func (h *HTTPGateway) FindOwnerByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	user, err := h.OwnerService.FindOwnerByID(id)
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

func (h *HTTPGateway) UpdateOwnerByID(ctx *fiber.Ctx) error {
	// Implementation for updating owner by ID
	return ctx.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "UpdateOwnerByID not implemented yet",
	})
}

func (h *HTTPGateway) DeleteOwnerByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing owner ID",
		})
	}

	deletedUser, err := h.OwnerService.DeleteOwnerByID(id)
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

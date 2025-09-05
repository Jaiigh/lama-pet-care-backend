package gateways

import (
	"github.com/gofiber/fiber/v2"
)

// GetOwnerByID handles GET /owner/:id requests
func (h *HTTPGateway) GetOwnerByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	user, err := h.OwnerService.GetOwnerByID(id)
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

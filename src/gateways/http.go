package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService  service.IAuthService
	OwnerService *service.OwnerService // Add OwnerService for owner operations
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService, owner *service.OwnerService) {
	gateway := &HTTPGateway{
		AuthService:  auth,
		OwnerService: owner,
	}

	GatewayUsers(*gateway, app)
}

// Handler for GET /owner/:id
func (g *HTTPGateway) GetOwnerByID(c *fiber.Ctx) error {
	id := c.Params("id")
	// You may want to validate id here
	owner, err := g.OwnerService.GetOwnerByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(owner)
}

package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService  service.IAuthService
	OwnerService *service.OwnerService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService, owner *service.OwnerService) {
	gateway := &HTTPGateway{
		AuthService:  auth,
		OwnerService: owner,
	}

	GatewayUsers(*gateway, app)
}

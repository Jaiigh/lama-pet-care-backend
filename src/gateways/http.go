package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService service.IAuthService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService) {
	gateway := &HTTPGateway{
		AuthService: auth,
	}

	GatewayUsers(*gateway, app)
}

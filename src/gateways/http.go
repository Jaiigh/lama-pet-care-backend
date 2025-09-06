package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService  service.IAuthService
	OwnerService *service.OwnerService
	AdminService *service.AdminService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService, owner *service.OwnerService, admin *service.AdminService) {
	gateway := &HTTPGateway{
		AuthService:  auth,
		OwnerService: owner,
		AdminService: admin,
	}

	GatewayUsers(*gateway, app)
}

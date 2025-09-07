package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService   service.IAuthService
	OwnerService  *service.OwnerService
	AdminService  *service.AdminService
	DoctorService *service.DoctorService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService,
	owner *service.OwnerService, admin *service.AdminService,
	doctor *service.DoctorService) {
	gateway := &HTTPGateway{
		AuthService:   auth,
		OwnerService:  owner,
		AdminService:  admin,
		DoctorService: doctor,
	}

	GatewayUsers(*gateway, app)
}

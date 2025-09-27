package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService      service.IAuthService
	OwnerService     service.IOwnerService
	AdminService     service.IAdminService
	DoctorService    service.IDoctorService
	CaretakerService service.ICaretakerService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService,
	owner service.IOwnerService, admin service.IAdminService,
	doctor service.IDoctorService, caretaker service.ICaretakerService) {
	gateway := &HTTPGateway{
		AuthService:      auth,
		OwnerService:     owner,
		AdminService:     admin,
		DoctorService:    doctor,
		CaretakerService: caretaker,
	}

	GatewayUsers(*gateway, app)
}

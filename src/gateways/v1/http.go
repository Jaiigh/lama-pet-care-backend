package gateways

import (
	service "lama-backend/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	AuthService      service.IAuthService
	UsersService     service.IUsersService
	OwnerService     service.IOwnerService
	DoctorService    service.IDoctorService
	CaretakerService service.ICaretakerService
	ServiceService	 service.IServiceService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService,
	users service.IUsersService, owner service.IOwnerService,
	doctor service.IDoctorService, caretaker service.ICaretakerService, service service.IServiceService) {
	gateway := &HTTPGateway{
		AuthService:      auth,
		UsersService:     users,
		OwnerService:     owner,
		DoctorService:    doctor,
		CaretakerService: caretaker,
		ServiceService:   service,
	}

	GatewayUsers(*gateway, app)
}

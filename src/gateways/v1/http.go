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
	ServiceService   service.IServiceService
	LeavedayService  service.ILeavedayService
	PetService       service.IPetService
	PaymentService   service.IPaymentService
}

func NewHTTPGateway(app *fiber.App, auth service.IAuthService,
	users service.IUsersService,
	owner service.IOwnerService, doctor service.IDoctorService, caretaker service.ICaretakerService,
	service service.IServiceService,
	leaveday service.ILeavedayService,
	pet service.IPetService,
	payment service.IPaymentService,) {
	gateway := &HTTPGateway{
		AuthService:      auth,
		UsersService:     users,
		OwnerService:     owner,
		DoctorService:    doctor,
		CaretakerService: caretaker,
		ServiceService:   service,
		LeavedayService:  leaveday,
		PetService:       pet,
		PaymentService:   payment,
	}

	GatewayUsers(*gateway, app)
}

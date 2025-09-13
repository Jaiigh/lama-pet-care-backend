package gateways

import (
	"lama-backend/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func GatewayUsers(gateway HTTPGateway, app *fiber.App) {
	// user := app.Group("/users")
	// user.Post("/create", gateway.CreateUser)
	// user.Get("/get_all", gateway.GetAllUserData)
	// user.Get("/get", gateway.GetByID)
	// user.Put("/update", gateway.UpdateUser)
	// user.Delete("/delete", gateway.DeleteUser)

	auth := app.Group("/auth")
	// check to login with token if not pass go to login with password
	auth.Get("/check_token", middlewares.SetJWtHeaderHandler(), gateway.checkToken)
	auth.Post("/register", gateway.Register)
	auth.Post("/login", gateway.Login)
	auth.Post("/create_admin", gateway.CreateAdmin)

	owner := app.Group("/owner")
	owner.Get("/", middlewares.SetJWtHeaderHandler(), gateway.FindOwnerByID)
	owner.Patch("/", middlewares.SetJWtHeaderHandler(), gateway.UpdateOwnerByID)
	owner.Delete("/", middlewares.SetJWtHeaderHandler(), gateway.DeleteOwnerByID)

	admin := app.Group("/admin")
	admin.Get("/", middlewares.SetJWtHeaderHandler(), gateway.FindAdminByID)
	admin.Patch("/", middlewares.SetJWtHeaderHandler(), gateway.UpdateAdminByID)
	admin.Delete("/", middlewares.SetJWtHeaderHandler(), gateway.DeleteAdminByID)

	doctor := app.Group("/doctor")
	doctor.Get("/", middlewares.SetJWtHeaderHandler(), gateway.FindDoctorByID)
	doctor.Patch("/", middlewares.SetJWtHeaderHandler(), gateway.UpdateDoctorByID)
	doctor.Delete("/", middlewares.SetJWtHeaderHandler(), gateway.DeleteDoctorByID)

	caretaker := app.Group("/caretaker")
	caretaker.Get("/", middlewares.SetJWtHeaderHandler(), gateway.FindCaretakerByID)
	caretaker.Patch("/", middlewares.SetJWtHeaderHandler(), gateway.UpdateCaretakerByID)
	caretaker.Delete("/", middlewares.SetJWtHeaderHandler(), gateway.DeleteCaretakerByID)
}

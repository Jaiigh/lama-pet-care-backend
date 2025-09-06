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
	owner.Get("/:id", gateway.FindOwnerByID)
	owner.Patch("/:id", gateway.UpdateOwnerByID)
	owner.Delete("/:id", gateway.DeleteOwnerByID)

	admin := app.Group("/admin")
	admin.Get("/:id", gateway.FindAdminByID)
	admin.Patch("/:id", gateway.UpdateAdminByID)
	admin.Delete("/:id", gateway.DeleteAdminByID)

	// doctor := app.Group("/doctor")
	// doctor.Get("/:id", gateway.FindDoctorByID)
	// doctor.Patch("/:id", gateway.UpdateDoctorByID)
	// doctor.Delete("/:id", gateway.DeleteDoctorByID)

	// caretaker := app.Group("/caretaker")
	// caretaker.Get("/:id", gateway.FindCaretakerByID)
	// caretaker.Patch("/:id", gateway.UpdateCaretakerByID)
	// caretaker.Delete("/:id", gateway.DeleteCaretakerByID)
}

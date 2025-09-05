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

	owner := app.Group("/owner")
	owner.Get("/:id", gateway.GetOwnerByID) // Handler to be implemented in HTTPGateway

	auth := app.Group("/auth")
	// check to login with token if not pass go to login with password
	auth.Get("/check_token", middlewares.SetJWtHeaderHandler(), gateway.checkToken)
	auth.Post("/register", gateway.Register)
	auth.Post("/login", gateway.Login)
	auth.Post("/create_admin", gateway.CreateAdmin)
}

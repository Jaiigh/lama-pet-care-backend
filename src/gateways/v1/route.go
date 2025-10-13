package gateways

import (
	"lama-backend/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func GatewayUsers(gateway HTTPGateway, app *fiber.App) {

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	// check to login with token if not pass go to login with password
	auth.Get("/token", middlewares.SetJWtHeaderHandler(), gateway.checkToken)
	auth.Post("/register/:role", gateway.Register)
	auth.Post("/login/:role", gateway.Login)
	auth.Post("/admin", gateway.CreateAdmin)

	user := api.Group("/user")
	user.Get("/", middlewares.SetJWtHeaderHandler(), gateway.FindUserByID)
	user.Patch("/", middlewares.SetJWtHeaderHandler(), gateway.UpdateUserByID)
	user.Delete("/", middlewares.SetJWtHeaderHandler(), gateway.DeleteUserByID)

	services := api.Group("/services", middlewares.SetJWtHeaderHandler())
	services.Post("/", gateway.CreateService)
	services.Delete("/:serviceID", gateway.DeleteService)
}

package gateways

import (
	"lama-backend/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func GatewayUsers(gateway HTTPGateway, app *fiber.App) {

	auth := app.Group("/auth")
	// check to login with token if not pass go to login with password
	auth.Get("/check_token", middlewares.SetJWtHeaderHandler(), gateway.checkToken)
	auth.Post("/register", gateway.Register)
	auth.Post("/login", gateway.Login)
	auth.Post("/create_admin", gateway.CreateAdmin)

	user := app.Group("/user")
	user.Get("/", middlewares.SetJWtHeaderHandler(), gateway.FindUserByID)
	user.Patch("/", middlewares.SetJWtHeaderHandler(), gateway.UpdateUserByID)
	user.Delete("/", middlewares.SetJWtHeaderHandler(), gateway.DeleteUserByID)
}

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
	auth.Post("/admin", middlewares.SetJWtHeaderHandler(), gateway.CreateAdmin)
	auth.Post("/password/email", gateway.ForgotPassword)
	auth.Patch("/password", gateway.ResetPassword)

	user := api.Group("/user", middlewares.SetJWtHeaderHandler())
	user.Get("/", gateway.FindUserByID)
	user.Patch("/", gateway.UpdateUserByID)
	user.Patch("/profile", gateway.UpdateUserPicture)
	user.Delete("/", gateway.DeleteUserByID)

	services := api.Group("/services", middlewares.SetJWtHeaderHandler())
	services.Post("/", gateway.CreateService)
	services.Get("/", gateway.GetMyServices)
	services.Patch("/:serviceID", gateway.UpdateService)
	services.Delete("/:serviceID", gateway.DeleteService)
}

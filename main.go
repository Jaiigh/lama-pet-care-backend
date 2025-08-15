package main

import (
	"lama-backend/configuration"
	ds "lama-backend/domain/datasources"
	repo "lama-backend/domain/repositories"
	gw "lama-backend/src/gateways"
	"lama-backend/src/middlewares"
	sv "lama-backend/src/services"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {

	// // // remove this before deploy ###################
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// /// ############################################

	app := fiber.New(configuration.NewFiberConfiguration())
	middlewares.Logger(app)
	app.Use(recover.New())
	app.Use(cors.New())

	prismadb := ds.ConnectPrisma()

	defer prismadb.PrismaDB.Prisma.Disconnect()

	userRepo := repo.NewUsersRepository(prismadb)

	sv0 := sv.NewUsersService(userRepo)

	gw.NewHTTPGateway(app, sv0)

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8080"
	}

	app.Listen(":" + PORT)
}

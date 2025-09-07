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

	adminRepo := repo.NewAdminRepository(prismadb)
	ownerRepo := repo.NewOwnerRepository(prismadb)
	caretakerRepo := repo.NewCaretakerRepository(prismadb)
	doctorRepo := repo.NewDoctorRepository(prismadb)

	svAuth := sv.NewAuthService(adminRepo, ownerRepo, caretakerRepo, doctorRepo)
	ownerService := sv.NewOwnerService(ownerRepo)
	adminService := sv.NewAdminService(adminRepo)
	doctorService := sv.NewDoctorService(doctorRepo)

	gw.NewHTTPGateway(app, svAuth, ownerService, adminService, doctorService)

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8080"
	}

	app.Listen(":" + PORT)
}

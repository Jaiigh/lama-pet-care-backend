package main

import (
	"lama-backend/configuration"
	ds "lama-backend/domain/datasources"
	repo "lama-backend/domain/repositories"
	gw "lama-backend/src/gateways/v1"
	"lama-backend/src/middlewares"
	sv "lama-backend/src/services"
	"log"
	"os"

	_ "lama-backend/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// @title LAMA Backend API
// @version 1.0
// @description this is a backend REST API server for LAMA project
// @host lama-pet-care-backend-dev.onrender.com
// @BasePath /api/v1
// @Schemes https
// @security BearerAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and then your JWT token.
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

	usersRepo := repo.NewUsersRepository(prismadb)
	ownerRepo := repo.NewOwnerRepository(prismadb)
	caretakerRepo := repo.NewCaretakerRepository(prismadb)
	doctorRepo := repo.NewDoctorRepository(prismadb)
	serviceRepo := repo.NewServiceRepository(prismadb)

	authService := sv.NewAuthService(usersRepo, ownerRepo, caretakerRepo, doctorRepo)
	usersService := sv.NewUsersService(usersRepo, ownerRepo, caretakerRepo, doctorRepo)
	ownerService := sv.NewOwnerService(ownerRepo)
	doctorService := sv.NewDoctorService(doctorRepo)
	caretakerService := sv.NewCaretakerService(caretakerRepo)
	serviceService := sv.NewServiceService(serviceRepo, caretakerRepo, doctorRepo)

	gw.NewHTTPGateway(app, authService, usersService, ownerService, doctorService, caretakerService, serviceService)

	app.Get("/swagger/*", swagger.HandlerDefault)

	PORT := os.Getenv("PORT")

	if PORT == "" {
		PORT = "8080"
	}

	app.Listen(":" + PORT)
}

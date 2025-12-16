package main

import (
	"fmt"
	"log"
	"os"

	"hifzhun-api/api/handlers"
	"hifzhun-api/api/routes"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	godotenv.Load()

	// Koneksi ke database
	config.ConnectDatabase()

	// Buat Fiber app
	app := fiber.New()

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins:     "http://localhost:5173",
	// 	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	// 	AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	// 	ExposeHeaders:    "Content-Length",
	// 	AllowCredentials: true,
	// }))

	userRepo := repositories.NewUserRepository(config.DB)
	authSvc := services.NewAuthService()
	authUC := usecases.NewAuthUsecase(userRepo, authSvc)
	authHandler := handlers.NewAuthHandler(authUC)

	routes.SetupRoutes(app, authHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on port %s...\n", port)
	app.Listen(fmt.Sprintf(":%s", port))
}

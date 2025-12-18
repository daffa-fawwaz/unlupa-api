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
	godotenv.Load()

	config.ConnectDatabase()

	app := fiber.New()

	// ================= REPOSITORY =================
	userRepo := repositories.NewUserRepository(config.DB)
	teacherReqRepo := repositories.NewTeacherRequestRepository(config.DB)

	// ================= AUTH =================
	authSvc := services.NewAuthService()
	authUC := usecases.NewAuthUsecase(userRepo, authSvc)
	authHandler := handlers.NewAuthHandler(authUC)

	// ================= USER (ADMIN) =================
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	// ================= TEACHER REQUEST =================
	teacherReqSvc := services.NewTeacherRequestService(teacherReqRepo, userRepo)
	teacherReqHandler := handlers.NewTeacherRequestHandler(teacherReqSvc)

	// ================= ROUTES =================
	routes.SetupRoutes(app, authHandler, userHandler, teacherReqHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on port %s...\n", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}


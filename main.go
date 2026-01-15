package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"hifzhun-api/api/handlers"
	"hifzhun-api/api/routes"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/seeders"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	seedFlag := flag.Bool("seed", false, "Run database seeders")
	flag.Parse()

	godotenv.Load()

	config.ConnectDatabase()

	if *seedFlag {
		log.Println("Running seeders...")
		ownerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
		if err := seeders.SeedFSRSWeights(config.DB); err != nil {
			log.Fatalf("Failed to seed FSRS weights: %v", err)
		}
		if err := seeders.SeedItems(config.DB, ownerID); err != nil {
			log.Fatalf("Failed to seed items: %v", err)
		}
		if err := seeders.SeedCards(config.DB); err != nil {
			log.Fatalf("Failed to seed cards: %v", err)
		}
		log.Println("Seeders completed successfully")
	}

	app := fiber.New()

	// ================= REPOSITORY =================
	userRepo := repositories.NewUserRepository(config.DB)
	teacherReqRepo := repositories.NewTeacherRequestRepository(config.DB)
	reviewStateRepo := repositories.NewReviewStateRepository(config.DB)
	cardRepo := repositories.NewCardRepository(config.DB)
	reviewLogRepo := repositories.NewReviewLogRepository(config.DB)
	fsrsWeightsRepo := repositories.NewFSRSWeightsRepository(config.DB)

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

	// ================= LOAD CONTROL =================
	loadControlSvc := services.NewLoadControlService(reviewStateRepo)
	loadControlHandler := handlers.NewLoadControlHandler(loadControlSvc)

	// ================= CARD / REVIEW =================
	reviewSvc := services.NewReviewService(
	config.DB,
	cardRepo,
	reviewLogRepo,
	reviewStateRepo,
	fsrsWeightsRepo,
)
	cardHandler := handlers.NewCardHandler(reviewSvc)

	// ================= DAILY TASK =================
	dailyTaskRepo := repositories.NewDailyTaskRepository(config.DB)
	dailyTaskSvc := services.NewDailyTaskService(
	reviewStateRepo,
	dailyTaskRepo,
)
    dailyTaskHandler := handlers.NewDailyTaskHandler(dailyTaskSvc)

	
dailyTaskActionRepo := repositories.NewDailyTaskActionRepository(config.DB)

dailyTaskActionSvc := services.NewDailyTaskActionService(
	dailyTaskActionRepo,
)

dailyTaskActionHandler := handlers.NewDailyTaskActionHandler(
	dailyTaskActionSvc,
)



	// ================= ROUTES =================
routes.SetupRoutes(
	app,
	authHandler,
	userHandler,
	teacherReqHandler,
	loadControlHandler,
	cardHandler,
	dailyTaskHandler,
	dailyTaskActionHandler, // 🔥 WAJIB
)


	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on port %s...\n", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}


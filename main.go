package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"hifzhun-api/api/handlers"
	"hifzhun-api/api/routes"
	_ "hifzhun-api/docs" // swagger docs
	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// @title Hifzhun API
// @version 1.0
// @description API untuk aplikasi hafalan Al-Quran dan Kitab
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@hifzhun.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	godotenv.Load()

	config.ConnectDatabase()
	config.InitRedis()
	appCache := cache.New(config.RedisClient)

	app := fiber.New()

	app.Use(cors.New(cors.Config{
	AllowOrigins:     "http://localhost:5173,http://localhost:3000",
	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	AllowCredentials: true,
}))

	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// ================= REPOSITORY =================
	userRepo := repositories.NewUserRepository(config.DB)
	teacherReqRepo := repositories.NewTeacherRequestRepository(config.DB)
	reviewStateRepo := repositories.NewReviewStateRepository(config.DB)
	fsrsWeightsRepo := repositories.NewFSRSWeightsRepository(config.DB)
	classRepo := repositories.NewClassRepository(config.DB)
	classMemberRepo := repositories.NewClassMemberRepository(config.DB)
	juzRepo := repositories.NewJuzRepository(config.DB)

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

	// ================= DAILY TASK =================
	// Note: itemRepo is declared here first for daily task service
	itemRepoForDaily := repositories.NewItemRepository(config.DB)
	juzItemRepo := repositories.NewJuzItemRepository(config.DB)
	dailyTaskRepo := repositories.NewDailyTaskRepository(config.DB)
	dailyTaskSvc := services.NewDailyTaskService(
		reviewStateRepo,
		dailyTaskRepo,
		itemRepoForDaily,
		classMemberRepo,
		classRepo,
		juzRepo,
	)
    dailyTaskHandler := handlers.NewDailyTaskHandler(dailyTaskSvc, itemRepoForDaily, juzItemRepo, appCache)


dailyTaskActionRepo := repositories.NewDailyTaskActionRepository(config.DB)



// graduation engine
graduationPreEngineRepo := repositories.NewItemGraduationRepository(config.DB)
graduationPreEngineSvc := services.NewGraduationPreEngine(graduationPreEngineRepo)
graduationPreEngineHandler := handlers.NewGraduationPreEngineHandler(graduationPreEngineSvc)


// ================= HAFALAN (JUZ & JUZ ITEM) =================
quranValidator, err := services.NewQuranValidator("data/surah.json")
if err != nil {
	log.Fatalf("Failed to initialize QuranValidator: %v", err)
}
itemRepo := repositories.NewItemRepository(config.DB)
hafalanSvc := services.NewHafalanService(juzRepo, itemRepo, juzItemRepo, quranValidator)
juzHandler := handlers.NewJuzHandler(hafalanSvc, juzRepo, juzItemRepo, appCache)
juzItemHandler := handlers.NewJuzItemHandler(hafalanSvc, appCache)

// ================= ITEM STATUS =================
intervalReviewLogRepo := repositories.NewIntervalReviewLogRepository(config.DB)
itemStatusSvc := services.NewItemStatusService(itemRepo, intervalReviewLogRepo)
itemStatusHandler := handlers.NewItemStatusHandler(itemStatusSvc)

// ================= BOOK =================
bookRepo := repositories.NewBookRepository(config.DB)
bookModuleRepo := repositories.NewBookModuleRepository(config.DB)
bookItemRepo := repositories.NewBookItemRepository(config.DB)
bookSvc := services.NewBookService(bookRepo, bookModuleRepo, bookItemRepo, itemRepo)
bookHandler := handlers.NewBookHandler(bookSvc, appCache)

// ================= CLASS =================
classBookRepo := repositories.NewClassBookRepository(config.DB)
classSvc := services.NewClassService(classRepo, classMemberRepo, classBookRepo, bookRepo, userRepo, itemRepo)
classHandler := handlers.NewClassHandler(classSvc)

// ================= ITEM REVIEW =================
itemReviewSvc := services.NewItemReviewService(itemRepo, fsrsWeightsRepo, dailyTaskActionRepo, classMemberRepo, classRepo)
itemReviewHandler := handlers.NewItemReviewHandler(itemReviewSvc, juzItemRepo, appCache)

// ================= MY ITEMS =================
myItemSvc := services.NewMyItemService(itemRepo, juzItemRepo, bookRepo, bookItemRepo)
myItemHandler := handlers.NewMyItemHandler(myItemSvc, appCache)

	// ================= ROUTES =================
routes.SetupRoutes(
	app,
	authHandler,
	userHandler,
	teacherReqHandler,
	loadControlHandler,
	dailyTaskHandler,
	graduationPreEngineHandler,
	juzHandler,
	juzItemHandler,
	itemStatusHandler,
	itemReviewHandler,
	bookHandler,
	classHandler,
	myItemHandler,
)


	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on port %s...\n", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

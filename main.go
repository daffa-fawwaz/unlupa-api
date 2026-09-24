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
	"hifzhun-api/pkg/middlewares"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/usecases"
	"hifzhun-api/pkg/utils"

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
	config.InitAppLocation()
	config.InitRedis()
	appCache := cache.New(config.RedisClient)

	app := fiber.New(fiber.Config{
		BodyLimit: utils.MaxImageSize + 1024*1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if fiberErr, ok := err.(*fiber.Error); ok {
				if fiberErr.Code == fiber.StatusRequestEntityTooLarge {
					return utils.Error(c, fiber.StatusBadRequest, "image size must be 3MB or less", "BAD_REQUEST", nil)
				}
				return utils.Error(c, fiberErr.Code, fiberErr.Message, "ERROR", nil)
			}

			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR", nil)
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:5174,https://unlupa.id,https://www.unlupa.id,https://api.unlupa.id",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Use(middlewares.ObservabilityMiddleware())

	// Serve uploaded files
	app.Static("/uploads", "./uploads")

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
	bookRepo := repositories.NewBookRepository(config.DB)

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
	classBookRepoForDaily := repositories.NewClassBookRepository(config.DB)
	dailyTaskSvc := services.NewDailyTaskService(
		reviewStateRepo,
		dailyTaskRepo,
		itemRepoForDaily,
		classMemberRepo,
		classRepo,
		juzRepo,
		juzItemRepo,
	)
	dailyTaskHandler := handlers.NewDailyTaskHandler(dailyTaskSvc, itemRepoForDaily, juzItemRepo, bookRepo, repositories.NewBookItemRepository(config.DB), classBookRepoForDaily, appCache)

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
	hafalanSvc := services.NewHafalanService(juzRepo, itemRepo, juzItemRepo, classRepo, classMemberRepo, quranValidator)
	juzHandler := handlers.NewJuzHandler(hafalanSvc, juzRepo, juzItemRepo, appCache)
	juzItemHandler := handlers.NewJuzItemHandler(hafalanSvc, appCache, itemRepo, juzItemRepo)

	// ================= BOOK =================
	bookModuleRepo := repositories.NewBookModuleRepository(config.DB)
	bookItemRepo := repositories.NewBookItemRepository(config.DB)
	bookUpdateRequestRepo := repositories.NewBookUpdateRequestRepository(config.DB)
	classBookRepo := repositories.NewClassBookRepository(config.DB)
	bookItemOverrideRepo := repositories.NewBookItemOverrideRepository(config.DB)
	bookSvc := services.NewBookService(bookRepo, bookModuleRepo, bookItemRepo, classBookRepo, itemRepo, userRepo, bookUpdateRequestRepo, bookItemOverrideRepo)
	bookHandler := handlers.NewBookHandler(bookSvc, userRepo, appCache)

	// ================= ITEM STATUS =================
	intervalReviewLogRepo := repositories.NewIntervalReviewLogRepository(config.DB)
	itemStatusSvc := services.NewItemStatusService(itemRepo, intervalReviewLogRepo, classBookRepo, dailyTaskActionRepo)
	itemStatusHandler := handlers.NewItemStatusHandler(itemStatusSvc, juzItemRepo, bookRepo, bookItemRepo, itemRepo, bookItemOverrideRepo, appCache)

	// ================= CLASS =================
	classSvc := services.NewClassService(classRepo, classMemberRepo, classBookRepo, bookRepo, userRepo, itemRepo, juzRepo, juzItemRepo, dailyTaskRepo, dailyTaskSvc)
	classHandler := handlers.NewClassHandler(classSvc)

	// ================= ITEM REVIEW =================
	itemReviewSvc := services.NewItemReviewService(itemRepo, fsrsWeightsRepo, dailyTaskActionRepo, classMemberRepo, classRepo, classBookRepo, juzItemRepo)
	itemReviewHandler := handlers.NewItemReviewHandler(itemReviewSvc, juzItemRepo, appCache)

	// ================= MY ITEMS =================
	myItemSvc := services.NewMyItemService(itemRepo, juzItemRepo, bookRepo, bookItemRepo, bookItemOverrideRepo)
	myItemHandler := handlers.NewMyItemHandler(myItemSvc, appCache)

	// ================= CLASS DAILY =================
	classDailyHandler := handlers.NewClassDailyHandler(
		dailyTaskSvc,
		dailyTaskRepo,
		itemRepo,
		juzRepo,
		juzItemRepo,
		classMemberRepo,
		classRepo,
		classBookRepo,
		bookRepo,
		bookItemRepo,
		appCache,
	)

	// ================= DASHBOARD STATS =================
	dashboardSvc := services.NewDashboardService(config.DB, itemRepo, juzItemRepo, bookItemRepo, bookRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc, appCache)

	// ================= QURAN CATALOG =================
	quranCatalogRepo := repositories.NewQuranCatalogRepository(config.DB)
	quranCatalogSvc := services.NewQuranCatalogService(quranCatalogRepo, itemRepo, juzRepo, juzItemRepo, config.DB)
	quranCatalogHandler := handlers.NewQuranCatalogHandler(quranCatalogSvc, appCache)

	// ================= QURAN PAGES PROGRESS & REVIEW =================
	quranPageSvc := services.NewQuranPageService(config.DB)
	quranPageHandler := handlers.NewQuranPageHandler(quranPageSvc, appCache)

	// ================= AI BUILDER (GEMINI) =================
	aiSvc := services.NewAIService()
	aiHandler := handlers.NewAIHandler(aiSvc)

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
		classDailyHandler,
		dashboardHandler,
		quranCatalogHandler,
		quranPageHandler,
		aiHandler,
	)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Server running on port %s...\n", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

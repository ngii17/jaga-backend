package main

import (
	"log"

	"jaga-backend/internal/config"
	"jaga-backend/internal/database"
	"jaga-backend/internal/handler"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/routes"
	"jaga-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 1. Baca .env
	cfg := config.Load()

	// 2. Sambung ke database (sekaligus AutoMigrate)
	db := database.Connect(cfg)

	// 3. Rakit Repository — Auth
	userRepo := repository.NewUserRepository(db)

	// 3b. Rakit Repository — Cek
	legalRepo := repository.NewLegalEntityRepository(db)
	threatRepo := repository.NewThreatEntityRepository(db)
	quizRepo := repository.NewQuizRepository(db)

	// 3c. Rakit Repository — Lapor
	reportRepo := repository.NewReportRepository(db)

	// 3d. Rakit Repository — Komunitas
	testimonialRepo := repository.NewTestimonialRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// 4. Rakit Service
	authService := service.NewAuthService(userRepo, cfg)
	checkService := service.NewCheckService(legalRepo, threatRepo, quizRepo)
	reportService := service.NewReportService(reportRepo, threatRepo, testimonialRepo, cfg)
	testimonialService := service.NewTestimonialService(testimonialRepo)
	chatService := service.NewChatService(chatRepo, checkService, reportService, cfg)

	// 5. Rakit Handler
	authHandler := handler.NewAuthHandler(authService)
	checkHandler := handler.NewCheckHandler(checkService)
	reportHandler := handler.NewReportHandler(reportService)
	testimonialHandler := handler.NewTestimonialHandler(testimonialService)
	chatHandler := handler.NewChatHandler(chatService)

	// 6. Siapkan Fiber app
	app := fiber.New()

	// 6b. Pasang CORS — sebelum routes didaftarkan
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// 6c. Daftarkan routes
	routes.RegisterAuthRoutes(app, cfg, authHandler, userRepo)
	routes.RegisterCheckRoutes(app, checkHandler)
	routes.RegisterReportRoutes(app, cfg, reportHandler, userRepo)
	routes.RegisterTestimonialRoutes(app, cfg, testimonialHandler, userRepo)
	routes.RegisterChatRoutes(app, chatHandler)

	// 7. Jalankan server
	log.Printf("Server jalan di port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

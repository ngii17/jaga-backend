package routes

import (
	"time"

	"jaga-backend/internal/config"
	"jaga-backend/internal/handler"
	"jaga-backend/internal/middleware"
	"jaga-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RegisterReportRoutes(app *fiber.App, cfg *config.Config, reportHandler *handler.ReportHandler, userRepo repository.UserRepository) {
	report := app.Group("/report")

	submitLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan pengiriman laporan, coba lagi beberapa saat lagi",
			})
		},
	})

	statusLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak permintaan, coba lagi beberapa saat lagi",
			})
		},
	})

	authenticatedLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak permintaan, coba lagi beberapa saat lagi",
			})
		},
	})

	// Publik
	report.Post("/submit", submitLimiter, reportHandler.SubmitReport)
	report.Get("/:id/status", statusLimiter, reportHandler.GetReportStatus)

	// Wajib login + izin can_approve_report
	report.Get("/", authenticatedLimiter, middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), reportHandler.GetReports)
	report.Get("/evidence/:evidenceId", authenticatedLimiter, middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), reportHandler.GetEvidence)
	report.Get("/:id/draft", authenticatedLimiter, middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), reportHandler.GetReportDraft)
	report.Post("/:id/verify", authenticatedLimiter, middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), reportHandler.VerifyReport)
	report.Get("/:id", authenticatedLimiter, middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), reportHandler.GetReportDetail)
}

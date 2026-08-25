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

func RegisterTestimonialRoutes(app *fiber.App, cfg *config.Config, testimonialHandler *handler.TestimonialHandler, userRepo repository.UserRepository) {
	testimonial := app.Group("/testimonials")

	publicLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak permintaan, coba lagi beberapa saat lagi",
			})
		},
	})

	// Publik
	testimonial.Get("/", publicLimiter, testimonialHandler.GetApprovedTestimonials)
	testimonial.Post("/submit", publicLimiter, testimonialHandler.SubmitTestimonial)
	testimonial.Post("/:id/upvote", publicLimiter, testimonialHandler.Upvote)

	// Wajib login + izin can_approve_report
	testimonial.Get("/pending", middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), testimonialHandler.GetPendingTestimonials)
	testimonial.Post("/:id/review", middleware.RequireAuth(cfg), middleware.RequireApproveReport(userRepo), testimonialHandler.ReviewTestimonial)
}

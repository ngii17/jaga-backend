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

func RegisterAuthRoutes(app *fiber.App, cfg *config.Config, authHandler *handler.AuthHandler, userRepo repository.UserRepository) {
	auth := app.Group("/auth")

	loginLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan login, coba lagi beberapa saat lagi",
			})
		},
	})

	otpLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan verifikasi OTP, coba lagi beberapa saat lagi",
			})
		},
	})

	auth.Post("/login", loginLimiter, authHandler.Login)
	auth.Post("/verify-otp", otpLimiter, authHandler.VerifyOTP)

	auth.Get("/me", middleware.RequireAuth(cfg), authHandler.Me)
	auth.Post("/users", middleware.RequireAuth(cfg), middleware.RequireManageUsers(userRepo), authHandler.RegisterVerifikator)
}

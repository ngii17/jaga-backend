package routes

import (
	"time"

	"jaga-backend/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RegisterCheckRoutes(app *fiber.App, checkHandler *handler.CheckHandler) {
	check := app.Group("/check")

	checkLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak permintaan, coba lagi beberapa saat lagi",
			})
		},
	})

	check.Get("/search", checkLimiter, checkHandler.Search)
	check.Get("/quiz/questions", checkLimiter, checkHandler.GetQuizQuestions)
	check.Post("/quiz/submit", checkLimiter, checkHandler.SubmitQuiz)
	app.Get("/check/threats", checkHandler.ListThreats)
}

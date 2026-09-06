package routes

import (
	"time"

	"jaga-backend/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RegisterChatRoutes(app *fiber.App, chatHandler *handler.ChatHandler) {
	chat := app.Group("/chat")

	chatLimiter := limiter.New(limiter.Config{
		Max:        15,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak pesan, tunggu sebentar ya",
			})
		},
	})

	chat.Post("/", chatLimiter, chatHandler.SendMessage)
}

package handler

import (
	"jaga-backend/internal/service"
	"jaga-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

type sendChatMessageRequest struct {
	Message string `json:"message"`
}

// POST /chat (publik)
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var req sendChatMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format permintaan tidak valid"})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Pesan tidak boleh kosong"})
	}

	fingerprint := utils.GenerateFingerprint(c.IP(), c.Get("User-Agent"))

	reply, err := h.chatService.SendMessage(fingerprint, req.Message, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"reply": reply},
	})
}

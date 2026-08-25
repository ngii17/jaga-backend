package handler

import (
	"strconv"

	"jaga-backend/internal/models"
	"jaga-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type CheckHandler struct {
	checkService service.CheckService
}

func NewCheckHandler(checkService service.CheckService) *CheckHandler {
	return &CheckHandler{checkService: checkService}
}

// GET /check/search?q=namaAplikasi
func (h *CheckHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Parameter pencarian 'q' wajib diisi",
		})
	}

	result, err := h.checkService.SearchName(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// GET /check/quiz/questions?scenario=pencegahan
func (h *CheckHandler) GetQuizQuestions(c *fiber.Ctx) error {
	scenario := models.QuizScenario(c.Query("scenario"))
	if scenario == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Parameter 'scenario' wajib diisi",
		})
	}

	questions, err := h.checkService.GetQuizQuestions(scenario)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    questions,
	})
}

type submitQuizRequest struct {
	Scenario models.QuizScenario  `json:"scenario"`
	Answers  []service.QuizAnswer `json:"answers"`
}

// POST /check/quiz/submit
func (h *CheckHandler) SubmitQuiz(c *fiber.Ctx) error {
	var req submitQuizRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format permintaan tidak valid",
		})
	}

	result, err := h.checkService.SubmitQuiz(req.Scenario, req.Answers)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

var _ = strconv.Itoa // placeholder, hapus kalau strconv tidak dipakai

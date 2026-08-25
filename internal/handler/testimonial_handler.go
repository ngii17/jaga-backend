package handler

import (
	"jaga-backend/internal/models"
	"jaga-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type TestimonialHandler struct {
	testimonialService service.TestimonialService
}

func NewTestimonialHandler(testimonialService service.TestimonialService) *TestimonialHandler {
	return &TestimonialHandler{testimonialService: testimonialService}
}

type submitTestimonialRequest struct {
	ReportID *uint  `json:"report_id"`
	Story    string `json:"story"`
}

// POST /testimonials/submit (publik)
func (h *TestimonialHandler) SubmitTestimonial(c *fiber.Ctx) error {
	var req submitTestimonialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format permintaan tidak valid"})
	}

	input := service.SubmitTestimonialInput{
		ReportID:  req.ReportID,
		Story:     req.Story,
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}

	testimonial, err := h.testimonialService.SubmitTestimonial(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Testimoni berhasil dikirim, menunggu peninjauan admin",
		"data":    fiber.Map{"testimonial_id": testimonial.ID},
	})
}

// GET /testimonials (publik)
func (h *TestimonialHandler) GetApprovedTestimonials(c *fiber.Ctx) error {
	testimonials, err := h.testimonialService.GetApprovedTestimonials()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": testimonials})
}

// POST /testimonials/:id/upvote (publik)
func (h *TestimonialHandler) Upvote(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID testimoni tidak valid"})
	}

	if err := h.testimonialService.Upvote(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Upvote berhasil"})
}

// GET /testimonials/pending (wajib login + izin can_approve_report)
func (h *TestimonialHandler) GetPendingTestimonials(c *fiber.Ctx) error {
	views, err := h.testimonialService.GetPendingTestimonials()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": views})
}

type reviewTestimonialRequest struct {
	Status models.TestimonialStatus `json:"status"`
	Story  string                   `json:"story"`
}

// POST /testimonials/:id/review (wajib login + izin can_approve_report)
func (h *TestimonialHandler) ReviewTestimonial(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID testimoni tidak valid"})
	}

	var req reviewTestimonialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format permintaan tidak valid"})
	}

	if err := h.testimonialService.ReviewTestimonial(uint(id), req.Status, req.Story); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Testimoni berhasil ditinjau"})
}

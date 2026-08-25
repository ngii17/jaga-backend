package handler

import (
	"net/http"

	"jaga-backend/internal/models"
	"jaga-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// POST /report/submit (multipart/form-data)
func (h *ReportHandler) SubmitReport(c *fiber.Ctx) error {
	isAnonymous := c.FormValue("is_anonymous") != "false"

	input := service.SubmitReportInput{
		Category:            models.ReportCategory(c.FormValue("category")),
		ReportedName:        c.FormValue("reported_name"),
		ReportedAccountOrWA: c.FormValue("reported_account_or_wa"),
		Description:         c.FormValue("description"),
		Severity:            models.ReportSeverity(c.FormValue("severity")),
		IsAnonymous:         isAnonymous,
		IPAddress:           c.IP(),
		UserAgent:           c.Get("User-Agent"),
	}

	if !isAnonymous {
		input.ContactName = c.FormValue("contact_name")
		input.ContactEmail = c.FormValue("contact_email")
		input.ContactPhone = c.FormValue("contact_phone")
	}

	form, err := c.MultipartForm()
	if err == nil && form.File["evidence"] != nil {
		for _, fileHeader := range form.File["evidence"] {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			buffer := make([]byte, fileHeader.Size)
			if _, err := file.Read(buffer); err != nil {
				file.Close()
				continue
			}
			file.Close()

			mimeType := http.DetectContentType(buffer)

			input.Evidences = append(input.Evidences, service.EvidenceFile{
				FileName: fileHeader.Filename,
				MimeType: mimeType,
				Data:     buffer,
			})
		}
	}

	report, err := h.reportService.SubmitReport(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Laporan berhasil dikirim",
		"data":    fiber.Map{"report_id": report.ID},
	})
}

type verifyReportRequest struct {
	Status models.ReportStatus `json:"status"`
	Note   string              `json:"note"`
}

// POST /report/:id/verify
func (h *ReportHandler) VerifyReport(c *fiber.Ctx) error {
	reportID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID laporan tidak valid"})
	}

	var req verifyReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format permintaan tidak valid"})
	}

	verifiedBy := c.Locals("userID").(uuid.UUID)

	if err := h.reportService.VerifyReport(uint(reportID), req.Status, verifiedBy, req.Note); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "Status laporan berhasil diperbarui"})
}

// GET /report?status=pending
func (h *ReportHandler) GetReports(c *fiber.Ctx) error {
	status := models.ReportStatus(c.Query("status"))

	reports, err := h.reportService.GetReports(status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": reports})
}

// GET /report/:id/status
func (h *ReportHandler) GetReportStatus(c *fiber.Ctx) error {
	reportID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID laporan tidak valid"})
	}

	report, err := h.reportService.GetReportStatus(uint(reportID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"report_id":     report.ID,
			"category":      report.Category,
			"reported_name": report.ReportedName,
			"status":        report.Status,
			"updated_at":    report.UpdatedAt,
		},
	})
}

// GET /report/evidence/:evidenceId (wajib login + izin can_approve_report)
func (h *ReportHandler) GetEvidence(c *fiber.Ctx) error {
	evidenceID, err := c.ParamsInt("evidenceId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bukti tidak valid"})
	}

	data, mimeType, err := h.reportService.GetDecryptedEvidence(uint(evidenceID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	c.Set("Content-Type", mimeType)
	return c.Send(data)
}

// GET /report/:id/draft (wajib login + izin can_approve_report)
func (h *ReportHandler) GetReportDraft(c *fiber.Ctx) error {
	reportID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID laporan tidak valid"})
	}

	draft, err := h.reportService.GenerateReportDraft(uint(reportID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": fiber.Map{"draft": draft}})
}

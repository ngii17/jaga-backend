package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"jaga-backend/internal/config"
	"jaga-backend/internal/models"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/utils"

	"github.com/google/uuid"
)

type EvidenceFile struct {
	FileName string
	MimeType string
	Data     []byte
}

type SubmitReportInput struct {
	Category            models.ReportCategory
	ReportedName        string
	ReportedAccountOrWA string
	Description         string
	Severity            models.ReportSeverity
	IsAnonymous         bool
	ContactName         string
	ContactEmail        string
	ContactPhone        string
	Evidences           []EvidenceFile
	IPAddress           string
	UserAgent           string
}

type reportService struct {
	reportRepo      repository.ReportRepository
	threatRepo      repository.ThreatEntityRepository
	testimonialRepo repository.TestimonialRepository
	cfg             *config.Config
}

type ReportService interface {
	SubmitReport(input SubmitReportInput) (*models.Report, error)
	VerifyReport(reportID uint, newStatus models.ReportStatus, verifiedBy uuid.UUID, note string) error
	GetReports(status models.ReportStatus) ([]models.Report, error)
	GetReportStatus(reportID uint) (*models.Report, error)
	GetDecryptedEvidence(evidenceID uint) ([]byte, string, error)
	GenerateReportDraft(reportID uint) (*ReportDraftResult, error)
	GetReportDetail(reportID uint) (*ReportDetailResult, error) // <- baris baru
}

type OfficialChannel struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

type ReportDraftResult struct {
	Draft            string            `json:"draft"`
	OfficialChannels []OfficialChannel `json:"official_channels"`
}

type ReportDetailResult struct {
	Report    *models.Report          `json:"report"`
	Evidences []models.ReportEvidence `json:"evidences"`
}

func (s *reportService) GetReportDetail(reportID uint) (*ReportDetailResult, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}

	evidences, err := s.reportRepo.FindEvidencesByReportID(reportID)
	if err != nil {
		return nil, errors.New("gagal mengambil data bukti")
	}

	return &ReportDetailResult{
		Report:    report,
		Evidences: evidences,
	}, nil
}

func NewReportService(reportRepo repository.ReportRepository, threatRepo repository.ThreatEntityRepository, testimonialRepo repository.TestimonialRepository, cfg *config.Config) ReportService {
	return &reportService{reportRepo: reportRepo, threatRepo: threatRepo, testimonialRepo: testimonialRepo, cfg: cfg}
}

const evidenceStorageDir = "storage/evidence"

func (s *reportService) SubmitReport(input SubmitReportInput) (*models.Report, error) {
	if !input.Category.IsValid() {
		return nil, errors.New("kategori laporan tidak valid")
	}
	if !input.Severity.IsValid() {
		return nil, errors.New("tingkat keparahan tidak valid")
	}

	fingerprint := utils.GenerateFingerprint(input.IPAddress, input.UserAgent)

	report := &models.Report{
		Category:             input.Category,
		ReportedName:         input.ReportedName,
		ReportedAccountOrWA:  input.ReportedAccountOrWA,
		Description:          input.Description,
		Severity:             input.Severity,
		IsAnonymous:          input.IsAnonymous,
		SubmitterFingerprint: fingerprint,
		Status:               models.ReportStatusPending,
	}

	var contact *models.ReportContact
	if !input.IsAnonymous {
		contact = &models.ReportContact{
			Name:  input.ContactName,
			Email: input.ContactEmail,
			Phone: input.ContactPhone,
		}
	}

	if err := os.MkdirAll(evidenceStorageDir, 0755); err != nil {
		return nil, errors.New("gagal menyiapkan folder penyimpanan bukti")
	}

	var evidences []models.ReportEvidence
	for _, ef := range input.Evidences {
		encrypted, err := utils.EncryptFile(s.cfg, ef.Data)
		if err != nil {
			return nil, errors.New("gagal mengenkripsi file bukti")
		}

		storedFileName := uuid.New().String() + ".enc"
		filePath := filepath.Join(evidenceStorageDir, storedFileName)

		if err := os.WriteFile(filePath, encrypted, 0600); err != nil {
			return nil, errors.New("gagal menyimpan file bukti")
		}

		evidences = append(evidences, models.ReportEvidence{
			FileName:         ef.FileName,
			FilePath:         filePath,
			DetectedMimeType: ef.MimeType,
			FileSize:         int64(len(ef.Data)),
		})
	}

	if err := s.reportRepo.CreateFull(report, contact, evidences); err != nil {
		return nil, fmt.Errorf("gagal menyimpan laporan: %w", err)
	}

	statusLog := &models.ReportStatusLog{
		ReportID:  report.ID,
		OldStatus: "",
		NewStatus: models.ReportStatusPending,
		ChangedBy: uuid.Nil,
		Note:      "Laporan baru masuk, menunggu verifikasi.",
	}
	if err := s.reportRepo.CreateStatusLog(statusLog); err != nil {
		return nil, errors.New("laporan tersimpan, tapi gagal mencatat log status")
	}

	return report, nil
}

func (s *reportService) VerifyReport(reportID uint, newStatus models.ReportStatus, verifiedBy uuid.UUID, note string) error {
	if !newStatus.IsValid() || newStatus == models.ReportStatusPending {
		return errors.New("status verifikasi tidak valid, harus 'diterima' atau 'ditolak'")
	}

	if newStatus != models.ReportStatusDiterima && newStatus != models.ReportStatusDitolak {
		return errors.New("status verifikasi tidak valid, harus 'diterima' atau 'ditolak'")
	}

	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return errors.New("laporan tidak ditemukan")
	}

	oldStatus := report.Status

	if err := s.reportRepo.UpdateStatus(reportID, newStatus); err != nil {
		return errors.New("gagal memperbarui status laporan")
	}

	statusLog := &models.ReportStatusLog{
		ReportID:  reportID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ChangedBy: verifiedBy,
		Note:      note,
	}
	if err := s.reportRepo.CreateStatusLog(statusLog); err != nil {
		return errors.New("status diperbarui, tapi gagal mencatat log")
	}

	if newStatus == models.ReportStatusDiterima {
		if err := s.threatRepo.IncrementOrCreate(report.ReportedName, report.Category); err != nil {
			return errors.New("status diperbarui, tapi gagal memperbarui data ancaman")
		}

		draftTestimonial := &models.Testimonial{
			ReportID: &report.ID,
			Story:    report.Description,
			Status:   models.TestimonialPending,
		}
		if err := s.testimonialRepo.Create(draftTestimonial); err != nil {
			return errors.New("status diperbarui, tapi gagal membuat draf testimoni")
		}
	}

	return nil
}

func (s *reportService) GetReports(status models.ReportStatus) ([]models.Report, error) {
	return s.reportRepo.FindAll(status)
}

func (s *reportService) GetReportStatus(reportID uint) (*models.Report, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}
	return report, nil
}

func (s *reportService) GetDecryptedEvidence(evidenceID uint) ([]byte, string, error) {
	evidence, err := s.reportRepo.FindEvidenceByID(evidenceID)
	if err != nil {
		return nil, "", errors.New("bukti tidak ditemukan")
	}

	encryptedData, err := os.ReadFile(evidence.FilePath)
	if err != nil {
		return nil, "", errors.New("gagal membaca file bukti")
	}

	decrypted, err := utils.DecryptFile(s.cfg, encryptedData)
	if err != nil {
		return nil, "", errors.New("gagal mendekripsi file bukti")
	}

	return decrypted, evidence.DetectedMimeType, nil
}

func (s *reportService) GenerateReportDraft(reportID uint) (*ReportDraftResult, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}

	evidences, err := s.reportRepo.FindEvidencesByReportID(reportID)
	if err != nil {
		return nil, errors.New("gagal mengambil data bukti")
	}

	draft := fmt.Sprintf(
		"LAPORAN DUGAAN %s\n"+
			"====================================\n\n"+
			"Nama/Entitas yang Dilaporkan : %s\n"+
			"Nomor Rekening/WhatsApp       : %s\n"+
			"Kategori                     : %s\n"+
			"Tingkat Keparahan            : %s\n"+
			"Tanggal Laporan Masuk        : %s\n\n"+
			"Kronologi Kejadian:\n%s\n\n"+
			"Jumlah Bukti Terlampir: %d file\n",
		report.Category, report.ReportedName, report.ReportedAccountOrWA,
		report.Category, report.Severity, report.CreatedAt.Format("02 January 2006, 15:04"),
		report.Description, len(evidences),
	)

	if len(evidences) > 0 {
		draft += "\nDaftar nama file bukti (lampirkan manual):\n"
		for i, e := range evidences {
			draft += fmt.Sprintf("%d. %s\n", i+1, e.FileName)
		}
	}

	return &ReportDraftResult{
		Draft:            draft,
		OfficialChannels: officialChannels,
	}, nil
}

// Data resmi, jarang berubah — makanya di-hardcode, bukan dari database.
var officialChannels = []OfficialChannel{
	{Name: "OJK - Layanan Konsumen", Contact: "Telepon 157 / WhatsApp 081-157-157-157"},
	{Name: "OJK - Email", Contact: "konsumen@ojk.go.id"},
	{Name: "Satgas Waspada Investasi", Contact: "aduan@waspadainvestasi.id"},
	{Name: "Kepolisian - Patroli Siber", Contact: "patrolisiber.id / info@cyber.polri.go.id"},
	{Name: "Kominfo - Aduan Konten", Contact: "aduankonten.id / aduankonten@kominfo.go.id / WA 08119224545"},
}

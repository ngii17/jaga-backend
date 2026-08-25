package service

import (
	"errors"

	"jaga-backend/internal/models"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/utils"
)

type SubmitTestimonialInput struct {
	ReportID  *uint
	Story     string
	IPAddress string
	UserAgent string
}

type PendingTestimonialView struct {
	Testimonial    models.Testimonial `json:"testimonial"`
	TotalSubmitted int                `json:"total_submitted"`
	TotalDitolak   int                `json:"total_ditolak"`
	IsMencurigakan bool               `json:"is_mencurigakan"`
}

type TestimonialService interface {
	SubmitTestimonial(input SubmitTestimonialInput) (*models.Testimonial, error)
	ReviewTestimonial(id uint, status models.TestimonialStatus, editedStory string) error
	GetApprovedTestimonials() ([]models.Testimonial, error)
	GetPendingTestimonials() ([]PendingTestimonialView, error)
	Upvote(id uint) error
}

type testimonialService struct {
	testimonialRepo repository.TestimonialRepository
}

func NewTestimonialService(testimonialRepo repository.TestimonialRepository) TestimonialService {
	return &testimonialService{testimonialRepo: testimonialRepo}
}

const mencurigakanThreshold = 2 // kalau sudah pernah ditolak 2x atau lebih, tandai mencurigakan

func (s *testimonialService) SubmitTestimonial(input SubmitTestimonialInput) (*models.Testimonial, error) {
	if input.Story == "" {
		return nil, errors.New("cerita testimoni tidak boleh kosong")
	}

	fingerprint := utils.GenerateFingerprint(input.IPAddress, input.UserAgent)

	testimonial := &models.Testimonial{
		ReportID:             input.ReportID,
		Story:                input.Story,
		Status:               models.TestimonialPending,
		SubmitterFingerprint: fingerprint,
	}

	if err := s.testimonialRepo.Create(testimonial); err != nil {
		return nil, errors.New("gagal mengirim testimoni, mungkin laporan ini sudah punya testimoni sebelumnya")
	}

	return testimonial, nil
}

func (s *testimonialService) ReviewTestimonial(id uint, status models.TestimonialStatus, editedStory string) error {
	if status != models.TestimonialDisetujui && status != models.TestimonialDitolak {
		return errors.New("status tidak valid, harus 'disetujui' atau 'ditolak'")
	}

	if _, err := s.testimonialRepo.FindByID(id); err != nil {
		return errors.New("testimoni tidak ditemukan")
	}

	return s.testimonialRepo.UpdateStatus(id, status, editedStory)
}

func (s *testimonialService) GetApprovedTestimonials() ([]models.Testimonial, error) {
	return s.testimonialRepo.FindApproved()
}

func (s *testimonialService) GetPendingTestimonials() ([]PendingTestimonialView, error) {
	pendingList, err := s.testimonialRepo.FindPending()
	if err != nil {
		return nil, errors.New("gagal mengambil daftar testimoni pending")
	}

	views := make([]PendingTestimonialView, 0, len(pendingList))
	for _, t := range pendingList {
		stats, err := s.testimonialRepo.GetFingerprintStats(t.SubmitterFingerprint)
		if err != nil {
			continue
		}

		views = append(views, PendingTestimonialView{
			Testimonial:    t,
			TotalSubmitted: stats.TotalSubmitted,
			TotalDitolak:   stats.TotalDitolak,
			IsMencurigakan: stats.TotalDitolak >= mencurigakanThreshold,
		})
	}

	return views, nil
}

func (s *testimonialService) Upvote(id uint) error {
	if _, err := s.testimonialRepo.FindByID(id); err != nil {
		return errors.New("testimoni tidak ditemukan")
	}
	return s.testimonialRepo.AddUpvote(id)
}

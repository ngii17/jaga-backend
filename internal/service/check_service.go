package service

import (
	"errors"

	"jaga-backend/internal/models"
	"jaga-backend/internal/repository"
)

type SearchStatus string

const (
	StatusAman                SearchStatus = "aman"
	StatusMiripWaspada        SearchStatus = "mirip_waspada"
	StatusTerkonfirmasiBahaya SearchStatus = "terkonfirmasi_bahaya"
)

type SearchResult struct {
	Status      SearchStatus `json:"status"`
	MatchedName string       `json:"matched_name,omitempty"`
	ReportCount int          `json:"report_count,omitempty"`
	Message     string       `json:"message"`
}

type QuizAnswer struct {
	QuestionID uint `json:"question_id"`
	Answer     bool `json:"answer"` // true = "Ya"
}

type QuizResult struct {
	TotalScore int    `json:"total_score"`
	LevelName  string `json:"level_name"`
	Message    string `json:"message"`
}

type CheckService interface {
	SearchName(name string) (*SearchResult, error)
	GetQuizQuestions(scenario models.QuizScenario) ([]models.QuizQuestion, error)
	SubmitQuiz(scenario models.QuizScenario, answers []QuizAnswer) (*QuizResult, error)
}

type checkService struct {
	legalRepo  repository.LegalEntityRepository
	threatRepo repository.ThreatEntityRepository
	quizRepo   repository.QuizRepository
}

func NewCheckService(
	legalRepo repository.LegalEntityRepository,
	threatRepo repository.ThreatEntityRepository,
	quizRepo repository.QuizRepository,
) CheckService {
	return &checkService{legalRepo: legalRepo, threatRepo: threatRepo, quizRepo: quizRepo}
}

const similarityThreshold = 0.3 // di bawah ini dianggap "tidak ada hasil relevan"

func (s *checkService) SearchName(name string) (*SearchResult, error) {
	// Cek dulu ke database bahaya - ini prioritas utama, warga harus tahu bahaya duluan.
	threats, err := s.threatRepo.SearchByName(name, 1)
	if err != nil {
		return nil, errors.New("gagal melakukan pencarian")
	}
	if len(threats) > 0 && threats[0].Name == name {
		return &SearchResult{
			Status:      StatusTerkonfirmasiBahaya,
			MatchedName: threats[0].Name,
			ReportCount: threats[0].ReportCount,
			Message:     "Aplikasi ini sudah dilaporkan dan diverifikasi berbahaya. Jangan gunakan.",
		}, nil
	}

	// Cek ke database legal.
	legals, err := s.legalRepo.SearchByName(name, 1)
	if err != nil {
		return nil, errors.New("gagal melakukan pencarian")
	}
	if len(legals) > 0 && legals[0].OfficialName == name {
		return &SearchResult{
			Status:      StatusAman,
			MatchedName: legals[0].OfficialName,
			Message:     "Terverifikasi resmi sebagai aplikasi legal. Aman digunakan.",
		}, nil
	}

	// Tidak identik di keduanya - cek apakah ada yang MIRIP.
	if len(legals) > 0 {
		return &SearchResult{
			Status:      StatusMiripWaspada,
			MatchedName: legals[0].OfficialName,
			Message:     "Nama ini mirip dengan aplikasi legal '" + legals[0].OfficialName + "', tapi bukan nama yang sama persis. Waspada, jangan install dulu.",
		}, nil
	}
	if len(threats) > 0 {
		return &SearchResult{
			Status:      StatusMiripWaspada,
			MatchedName: threats[0].Name,
			Message:     "Nama ini mirip dengan entitas yang pernah dilaporkan berbahaya. Waspada.",
		}, nil
	}

	return &SearchResult{
		Status:  StatusMiripWaspada,
		Message: "Belum ada data terkait nama ini. Tetap waspada dan cek ciri-cirinya lewat kuis di bawah.",
	}, nil
}

func (s *checkService) GetQuizQuestions(scenario models.QuizScenario) ([]models.QuizQuestion, error) {
	return s.quizRepo.GetQuestionsByScenario(scenario)
}

func (s *checkService) SubmitQuiz(scenario models.QuizScenario, answers []QuizAnswer) (*QuizResult, error) {
	totalScore := 0

	questions, err := s.quizRepo.GetQuestionsByScenario(scenario)
	if err != nil {
		return nil, errors.New("gagal mengambil data pertanyaan")
	}

	// Bikin "kamus" bobot per pertanyaan, supaya gampang dicocokkan dengan jawaban user.
	weightMap := make(map[uint]int)
	for _, q := range questions {
		weightMap[q.ID] = q.Weight
	}

	for _, ans := range answers {
		if ans.Answer {
			totalScore += weightMap[ans.QuestionID]
		}
	}

	level, err := s.quizRepo.GetRiskLevelByScore(scenario, totalScore)
	if err != nil {
		return nil, errors.New("skor tidak cocok dengan level risiko manapun")
	}

	return &QuizResult{
		TotalScore: totalScore,
		LevelName:  level.LevelName,
		Message:    level.Description,
	}, nil
}

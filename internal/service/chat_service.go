package service

import (
	"errors"
	"log"

	"jaga-backend/internal/config"
	"jaga-backend/internal/models"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/utils"
)

const maxUsagePerSession = 10
const maxFunctionCallLoops = 5 // pengaman, cegah Gemini "muter-muter" tanpa henti

type ChatService interface {
	SendMessage(fingerprint string, userMessage string, ipAddress string, userAgent string) (string, error)
}

type chatService struct {
	chatRepo      repository.ChatRepository
	checkService  CheckService
	reportService ReportService
	cfg           *config.Config
}

func NewChatService(chatRepo repository.ChatRepository, checkService CheckService, reportService ReportService, cfg *config.Config) ChatService {
	return &chatService{chatRepo: chatRepo, checkService: checkService, reportService: reportService, cfg: cfg}
}

func (s *chatService) SendMessage(fingerprint string, userMessage string, ipAddress string, userAgent string) (string, error) {
	// 1. Cari atau buat sesi
	session, err := s.chatRepo.FindSessionByFingerprint(fingerprint)
	if err != nil {
		session, err = s.chatRepo.CreateSession(fingerprint)
		if err != nil {
			return "", errors.New("gagal membuat sesi percakapan")
		}
	}

	// 2. Cek limiter
	if session.UsageCount >= maxUsagePerSession {
		return "Kamu sudah mencapai batas percakapan hari ini. Yuk langsung coba fitur Cek atau Lapor kami di menu utama untuk bantuan lebih lanjut.", nil
	}

	// 3. Ambil riwayat sebelumnya, susun jadi format Gemini
	previousMessages, err := s.chatRepo.GetMessagesBySession(session.ID)
	if err != nil {
		return "", errors.New("gagal mengambil riwayat percakapan")
	}

	history := make([]utils.GeminiContent, 0, len(previousMessages)+1)
	for _, msg := range previousMessages {
		geminiRole := "user"
		if msg.Role == models.ChatRoleAssistant {
			geminiRole = "model"
		}
		history = append(history, utils.GeminiContent{
			Role:  geminiRole,
			Parts: []utils.GeminiPart{{Text: msg.Content}},
		})
	}

	// 4. Simpan & tambahkan pesan baru dari user
	if err := s.chatRepo.SaveMessage(session.ID, models.ChatRoleUser, userMessage); err != nil {
		return "", errors.New("gagal menyimpan pesan")
	}
	history = append(history, utils.GeminiContent{
		Role:  "user",
		Parts: []utils.GeminiPart{{Text: userMessage}},
	})

	// 5. Loop: kirim ke Gemini, tangani kalau dia minta panggil fungsi
	var finalText string
	for i := 0; i < maxFunctionCallLoops; i++ {
		result, err := utils.SendChatToGemini(s.cfg, utils.JagaSystemPrompt, history)
		if err != nil {
			log.Printf("ERROR Gemini: %v", err)
			return "", errors.New("gagal berkomunikasi dengan asisten")
		}

		hasFunctionCall := false
		for _, part := range result.Parts {
			if part.FunctionCall != nil {
				hasFunctionCall = true

				// Jalankan function Go asli sesuai permintaan Gemini
				functionResult := s.executeFunction(part.FunctionCall, ipAddress, userAgent)

				// Catat "giliran" Gemini minta fungsi + hasil yang kita berikan balik
				history = append(history, utils.GeminiContent{
					Role: "model",
					Parts: []utils.GeminiPart{{
						FunctionCall:     part.FunctionCall,
						ThoughtSignature: part.ThoughtSignature,
					}},
				})
				history = append(history, utils.GeminiContent{
					Role: "user",
					Parts: []utils.GeminiPart{{
						FunctionResponse: &utils.GeminiFunctionResponse{
							Name:     part.FunctionCall.Name,
							Response: functionResult,
						},
					}},
				})
			} else if part.Text != "" {
				finalText = part.Text
			}
		}

		if !hasFunctionCall {
			break // Gemini sudah kasih jawaban final, tidak minta fungsi lagi
		}
	}

	if finalText == "" {
		finalText = "Maaf, aku belum bisa memproses itu. Bisa dicoba dengan pertanyaan lain?"
	}

	// 6. Simpan jawaban final, tambah usage count
	if err := s.chatRepo.SaveMessage(session.ID, models.ChatRoleAssistant, finalText); err != nil {
		return "", errors.New("jawaban berhasil dibuat, tapi gagal disimpan")
	}
	if err := s.chatRepo.IncrementUsage(session.ID); err != nil {
		return "", errors.New("gagal memperbarui batas pemakaian")
	}

	return finalText, nil
}

// executeFunction adalah "meja resepsionis" yang mencocokkan nama fungsi
// yang diminta Gemini ke function Go asli yang sesuai.
func (s *chatService) executeFunction(call *utils.GeminiFunctionCall, ipAddress string, userAgent string) map[string]interface{} {
	switch call.Name {
	case "cari_entitas":
		nama, _ := call.Args["nama"].(string)
		result, err := s.checkService.SearchName(nama)
		if err != nil {
			return map[string]interface{}{"error": "gagal melakukan pencarian"}
		}
		return map[string]interface{}{
			"status":       result.Status,
			"matched_name": result.MatchedName,
			"report_count": result.ReportCount,
			"message":      result.Message,
		}

	case "ambil_pertanyaan_kuis":
		skenario, _ := call.Args["skenario"].(string)
		questions, err := s.checkService.GetQuizQuestions(models.QuizScenario(skenario))
		if err != nil {
			return map[string]interface{}{"error": "gagal mengambil pertanyaan"}
		}
		log.Printf("DEBUG: skenario '%s' dapat %d pertanyaan dari database", skenario, len(questions))
		return map[string]interface{}{"questions": questions}

	case "submit_jawaban_kuis":
		skenario, _ := call.Args["skenario"].(string)
		idsRaw, _ := call.Args["question_ids"].([]interface{})

		var answers []QuizAnswer
		for _, idRaw := range idsRaw {
			idFloat, _ := idRaw.(float64) // JSON angka selalu terbaca sebagai float64 di Go
			answers = append(answers, QuizAnswer{QuestionID: uint(idFloat), Answer: true})
		}

		result, err := s.checkService.SubmitQuiz(models.QuizScenario(skenario), answers)
		if err != nil {
			return map[string]interface{}{"error": "gagal menghitung skor kuis"}
		}
		return map[string]interface{}{
			"total_score": result.TotalScore,
			"level_name":  result.LevelName,
			"message":     result.Message,
		}

	case "submit_laporan":
		category, _ := call.Args["category"].(string)
		reportedName, _ := call.Args["reported_name"].(string)
		description, _ := call.Args["description"].(string)
		severity, _ := call.Args["severity"].(string)
		isAnonymous, _ := call.Args["is_anonymous"].(bool)

		input := SubmitReportInput{
			Category:     models.ReportCategory(category),
			ReportedName: reportedName,
			Description:  description,
			Severity:     models.ReportSeverity(severity),
			IsAnonymous:  isAnonymous,
			IPAddress:    ipAddress,
			UserAgent:    userAgent,
		}

		report, err := s.reportService.SubmitReport(input)
		if err != nil {
			return map[string]interface{}{"error": err.Error()}
		}
		return map[string]interface{}{"report_id": report.ID, "status": "berhasil disubmit"}

	case "cek_status_laporan":
		idFloat, _ := call.Args["report_id"].(float64)
		report, err := s.reportService.GetReportStatus(uint(idFloat))
		if err != nil {
			return map[string]interface{}{"error": "laporan tidak ditemukan"}
		}
		return map[string]interface{}{"status": report.Status}

	default:
		return map[string]interface{}{"error": "fungsi tidak dikenali"}
	}
}

package main

import (
	"log"
	"time"

	"jaga-backend/internal/config"
	"jaga-backend/internal/database"
	"jaga-backend/internal/models"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/utils"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	userRepo := repository.NewUserRepository(db)

	email := "tonggisim28@gmail.com"
	rawPassword := "adminn123"

	// 1. SEEDING ADMIN (Diperbaiki agar tidak menghentikan seluruh proses)
	if _, err := userRepo.FindByEmail(email); err == nil {
		log.Println("Admin dengan email ini sudah ada, skip pembuatan admin.")
	} else {
		hash, err := utils.HashPassword(rawPassword)
		if err != nil {
			log.Fatalf("Gagal hash password: %v", err)
		}

		admin := &models.User{
			Name:             "Admin JAGA",
			Email:            email,
			PasswordHash:     hash,
			Role:             models.RoleAdmin,
			CanApproveReport: true,
			CanDeleteReport:  true,
			CanManageUsers:   true,
		}

		if err := userRepo.Create(admin); err != nil {
			log.Fatalf("Gagal membuat admin: %v", err)
		}
		log.Printf("Admin berhasil dibuat: %s / %s (SEGERA GANTI PASSWORD INI)", email, rawPassword)
	}

	// 2. SEEDING LEGAL ENTITIES (Menggunakan FirstOrCreate berdasarkan Nomor Izin OJK)
	legalEntities := []models.LegalEntity{
		{OfficialName: "Kredit Pintar", Category: models.CategoryPinjol, OJKLicenseNumber: "KEP-123/2021", Status: "legal"},
		{OfficialName: "Akulaku Finance", Category: models.CategoryPinjol, OJKLicenseNumber: "KEP-456/2020", Status: "legal"},
		{OfficialName: "Kredivo", Category: models.CategoryPinjol, OJKLicenseNumber: "KEP-789/2019", Status: "legal"},
		{OfficialName: "Bareksa", Category: models.CategoryInvestasi, OJKLicenseNumber: "KEP-321/2018", Status: "legal"},
	}

	for _, e := range legalEntities {
		// Mencegah duplikasi berdasarkan OJKLicenseNumber
		if err := db.Where(models.LegalEntity{OJKLicenseNumber: e.OJKLicenseNumber}).FirstOrCreate(&e).Error; err != nil {
			log.Printf("Gagal seeder legal entity %s: %v", e.OfficialName, err)
		}
	}

	// 3. SEEDING THREAT ENTITIES
	now := time.Now()
	threatEntities := []models.ThreatEntity{
		{Name: "DanaCepatt Ilegal", Category: models.CategoryPinjolIlegal, ReportCount: 12, FirstReportedAt: now.AddDate(0, -2, 0), LastReportedAt: now},
		{Name: "Slot Gacor88", Category: models.CategoryJudiOnline, ReportCount: 27, FirstReportedAt: now.AddDate(0, -1, 0), LastReportedAt: now},
		{Name: "Investasi Untung Cepat", Category: models.CategoryInvestasiBodong, ReportCount: 8, FirstReportedAt: now.AddDate(0, -3, 0), LastReportedAt: now},
	}

	for _, e := range threatEntities {
		// Mencegah duplikasi berdasarkan Nama Entity Bahaya
		if err := db.Where(models.ThreatEntity{Name: e.Name}).FirstOrCreate(&e).Error; err != nil {
			log.Printf("Gagal seeder threat entity %s: %v", e.Name, err)
		}
	}

	// 4. SEEDING QUIZ QUESTIONS (Pencegahan & Penanganan)
	quizQuestions := []models.QuizQuestion{
		// Skenario Pencegahan
		{Scenario: models.ScenarioPencegahan, QuestionText: "Aplikasi ini meminta akses ke seluruh kontak HP Anda?", Weight: 25, OrderIndex: 1},
		{Scenario: models.ScenarioPencegahan, QuestionText: "Aplikasi ini menjanjikan pencairan dana dalam hitungan menit tanpa verifikasi?", Weight: 20, OrderIndex: 2},
		{Scenario: models.ScenarioPencegahan, QuestionText: "Bunga atau biaya admin tidak dijelaskan secara rinci di awal?", Weight: 20, OrderIndex: 3},
		{Scenario: models.ScenarioPencegahan, QuestionText: "Aplikasi ini tidak terdaftar di OJK/tidak ada nomor izin?", Weight: 35, OrderIndex: 4},

		// Skenario Penanganan
		{Scenario: models.ScenarioPenanganan, QuestionText: "Anda sudah mengirimkan data pribadi (KTP/foto selfie) ke pihak tersebut?", Weight: 30, OrderIndex: 1},
		{Scenario: models.ScenarioPenanganan, QuestionText: "Anda sudah pernah transfer sejumlah uang ke rekening yang diminta?", Weight: 30, OrderIndex: 2},
		{Scenario: models.ScenarioPenanganan, QuestionText: "Anda mendapat ancaman atau teror dari pihak penagih?", Weight: 40, OrderIndex: 3},
	}

	for _, q := range quizQuestions {
		// Mencegah duplikasi berdasarkan teks pertanyaan dan skenario
		if err := db.Where(models.QuizQuestion{Scenario: q.Scenario, QuestionText: q.QuestionText}).FirstOrCreate(&q).Error; err != nil {
			log.Printf("Gagal seeder quiz question [%s]: %v", q.QuestionText, err)
		}
	}

	// 5. SEEDING RISK LEVELS
	riskLevels := []models.RiskLevel{
		{Scenario: models.ScenarioPencegahan, LevelName: "Aman", MinScore: 0, MaxScore: 30, Description: "Ciri-ciri yang Anda temukan minim indikasi bahaya. Tetap perhatikan legalitas OJK."},
		{Scenario: models.ScenarioPencegahan, LevelName: "Waspada", MinScore: 31, MaxScore: 60, Description: "Beberapa ciri mencurigakan ditemukan. Jangan lanjutkan sebelum memastikan legalitasnya."},
		{Scenario: models.ScenarioPencegahan, LevelName: "Bahaya", MinScore: 61, MaxScore: 100, Description: "Indikasi kuat ini adalah pinjol/judi ilegal. Jangan gunakan aplikasi ini."},
		{Scenario: models.ScenarioPenanganan, LevelName: "Perlu Pendampingan Ringan", MinScore: 0, MaxScore: 30, Description: "Segera hentikan komunikasi dan laporkan lewat Modul Lapor."},
		{Scenario: models.ScenarioPenanganan, LevelName: "Perlu Tindakan Segera", MinScore: 31, MaxScore: 70, Description: "Data pribadi/dana Anda berisiko. Laporkan sekarang dan hubungi pihak berwenang."},
		{Scenario: models.ScenarioPenanganan, LevelName: "Darurat", MinScore: 71, MaxScore: 100, Description: "Anda mengalami ancaman/teror. Segera laporkan ke polisi dan gunakan Modul Lapor."},
	}

	for _, r := range riskLevels {
		// Mencegah duplikasi berdasarkan Skenario dan Nama Level Risk
		if err := db.Where(models.RiskLevel{Scenario: r.Scenario, LevelName: r.LevelName}).FirstOrCreate(&r).Error; err != nil {
			log.Printf("Gagal seeder risk level %s: %v", r.LevelName, err)
		}
	}

	log.Println("Seeding Modul Cek selesai dengan aman: legal entities, threat entities, quiz questions, risk levels.")
}

package models

type QuizScenario string

const (
	ScenarioPencegahan QuizScenario = "pencegahan"
	ScenarioPenanganan QuizScenario = "penanganan"
)

// ReportCategory didefinisikan di sini karena dipakai bersama oleh
// ThreatEntity (Modul 2) dan Report (Modul 3 nanti).
type ReportCategory string

const (
	CategoryPinjolIlegal    ReportCategory = "pinjol_ilegal"
	CategoryJudiOnline      ReportCategory = "judi_online"
	CategoryInvestasiBodong ReportCategory = "investasi_bodong"
	CategoryReportLainnya   ReportCategory = "lainnya"
)

type QuizQuestion struct {
	ID           uint         `gorm:"primaryKey"`
	Scenario     QuizScenario `gorm:"type:varchar(20);not null"`
	QuestionText string       `gorm:"not null"`
	Weight       int          `gorm:"not null"`
	OrderIndex   int          `gorm:"default:0"`
}

func (QuizQuestion) TableName() string {
	return "quiz_questions"
}

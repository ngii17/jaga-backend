package models

type RiskLevel struct {
	ID          uint         `gorm:"primaryKey"`
	Scenario    QuizScenario `gorm:"type:varchar(20);not null"`
	LevelName   string       `gorm:"not null"`
	MinScore    int          `gorm:"not null"`
	MaxScore    int          `gorm:"not null"`
	Description string
}

func (RiskLevel) TableName() string {
	return "risk_levels"
}

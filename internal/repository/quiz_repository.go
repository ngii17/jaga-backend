package repository

import (
	"jaga-backend/internal/models"

	"gorm.io/gorm"
)

type QuizRepository interface {
	GetQuestionsByScenario(scenario models.QuizScenario) ([]models.QuizQuestion, error)
	GetRiskLevelByScore(scenario models.QuizScenario, score int) (*models.RiskLevel, error)
}

type quizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) QuizRepository {
	return &quizRepository{db: db}
}

func (r *quizRepository) GetQuestionsByScenario(scenario models.QuizScenario) ([]models.QuizQuestion, error) {
	var questions []models.QuizQuestion
	err := r.db.Where("scenario = ?", scenario).Order("order_index").Find(&questions).Error
	return questions, err
}

func (r *quizRepository) GetRiskLevelByScore(scenario models.QuizScenario, score int) (*models.RiskLevel, error) {
	var level models.RiskLevel
	err := r.db.Where("scenario = ? AND min_score <= ? AND max_score >= ?", scenario, score, score).First(&level).Error
	return &level, err
}

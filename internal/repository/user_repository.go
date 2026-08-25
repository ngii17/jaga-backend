package repository

import (
	"jaga-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository adalah "daftar janji" — nama fungsi yang WAJIB ada,
// tapi belum ada isi logic-nya di sini.
type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
}

// userRepository adalah implementasi ASLI dari interface di atas.
// Huruf kecil di awal ("u") sengaja, artinya struct ini
// tidak bisa diakses langsung dari luar package - lihat penjelasan di bawah.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository adalah "pabrik" untuk bikin userRepository baru.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	return &user, result.Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.Where("id = ?", id).First(&user)
	return &user, result.Error
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

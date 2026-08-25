package service

import (
	"errors"
	"time"

	"jaga-backend/internal/config"
	"jaga-backend/internal/models"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/utils"

	"github.com/google/uuid"
)

const (
	otpValidityMinutes = 5
	maxOTPAttempts     = 5
)

// AuthService adalah "daftar janji" - fungsi apa saja yang harus ada.
type AuthService interface {
	Login(email, password string) error
	VerifyOTP(email, otpCode string) (string, *models.User, error)
	RegisterVerifikator(name, email, password string, createdBy uuid.UUID) (*models.User, error)
	GetProfile(userID uuid.UUID) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{userRepo: userRepo, cfg: cfg}
}

// Login = tahap pertama. Kalau email+password valid, kirim OTP ke email.
func (s *authService) Login(email, password string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return errors.New("email atau password salah")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return errors.New("email atau password salah")
	}

	otpCode, err := utils.GenerateOTP()
	if err != nil {
		return errors.New("gagal membuat kode OTP")
	}

	otpHash, err := utils.HashOTP(otpCode)
	if err != nil {
		return errors.New("gagal memproses kode OTP")
	}

	expiresAt := time.Now().Add(otpValidityMinutes * time.Minute)
	user.OTPCodeHash = &otpHash
	user.OTPExpiresAt = &expiresAt
	user.OTPAttempts = 0

	if err := s.userRepo.Update(user); err != nil {
		return errors.New("gagal menyimpan kode OTP")
	}

	if err := utils.SendOTPEmail(s.cfg, user.Email, otpCode); err != nil {
		return errors.New("gagal mengirim email OTP, coba lagi")
	}

	return nil
}

// VerifyOTP = tahap kedua. Kalau kode benar & belum kedaluwarsa, terbitkan JWT.
func (s *authService) VerifyOTP(email, otpCode string) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, errors.New("email tidak ditemukan")
	}

	if user.OTPCodeHash == nil || user.OTPExpiresAt == nil {
		return "", nil, errors.New("belum ada permintaan OTP aktif, silakan login ulang")
	}

	if user.OTPAttempts >= maxOTPAttempts {
		return "", nil, errors.New("terlalu banyak percobaan salah, silakan login ulang")
	}

	if time.Now().After(*user.OTPExpiresAt) {
		return "", nil, errors.New("kode OTP sudah kedaluwarsa, silakan login ulang")
	}

	if !utils.CheckOTPHash(otpCode, *user.OTPCodeHash) {
		user.OTPAttempts++
		_ = s.userRepo.Update(user)
		return "", nil, errors.New("kode OTP salah")
	}

	user.OTPCodeHash = nil
	user.OTPExpiresAt = nil
	user.OTPAttempts = 0
	if err := s.userRepo.Update(user); err != nil {
		return "", nil, errors.New("gagal menyelesaikan proses login")
	}

	token, err := utils.GenerateJWT(s.cfg.JWTSecret, user.ID, string(user.Role))
	if err != nil {
		return "", nil, errors.New("gagal membuat sesi login")
	}

	return token, user, nil
}

// RegisterVerifikator hanya boleh dipanggil admin (dicek nanti di Handler/Middleware).
func (s *authService) RegisterVerifikator(name, email, password string, createdBy uuid.UUID) (*models.User, error) {
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, errors.New("email sudah terdaftar")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("gagal memproses password")
	}

	newUser := &models.User{
		Name:             name,
		Email:            email,
		PasswordHash:     hash,
		Role:             models.RoleVerifikator,
		CanApproveReport: true,
		CreatedBy:        &createdBy,
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, errors.New("gagal menyimpan akun baru")
	}
	return newUser, nil
}

// GetProfile mengambil data user yang sedang login, dipakai endpoint /auth/me.
func (s *authService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}
	return user, nil
}

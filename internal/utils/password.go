package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword mengubah password mentah jadi hash bcrypt.
// bcrypt.DefaultCost saat ini bernilai 10 — cukup aman, tidak terlalu lambat.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash membandingkan password mentah yang diketik user
// dengan hash yang tersimpan di database. Mengembalikan true kalau cocok.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

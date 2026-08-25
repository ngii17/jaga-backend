package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP menghasilkan kode 6 digit acak (000000-999999),
// aman secara kriptografis (tidak bisa ditebak polanya).
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashOTP dan CheckOTPHash memakai bcrypt yang sama seperti password,
// supaya kode OTP mentah tidak pernah tersimpan langsung di database.
func HashOTP(otp string) (string, error) {
	return HashPassword(otp)
}

func CheckOTPHash(otp, hash string) bool {
	return CheckPasswordHash(otp, hash)
}

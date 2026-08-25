package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// GenerateFingerprint membuat hash SHA-256 dari kombinasi IP address dan User-Agent.
// Dipakai untuk mendeteksi laporan berulang dari sumber yang sama, tanpa menyimpan IP asli.
func GenerateFingerprint(ip string, userAgent string) string {
	combined := ip + "|" + userAgent
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

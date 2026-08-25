package utils

import (
	"fmt"
	"net/smtp"

	"jaga-backend/internal/config"
)

// SendOTPEmail mengirim kode OTP ke email tujuan lewat Gmail SMTP.
func SendOTPEmail(cfg *config.Config, toEmail string, otpCode string) error {
	subject := "Kode OTP Login - JAGA"
	body := fmt.Sprintf(
		"Kode OTP Anda: %s\n\nKode ini berlaku 5 menit. Jangan bagikan ke siapapun.",
		otpCode,
	)

	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")

	auth := smtp.PlainAuth("", cfg.SMTPEmail, cfg.SMTPAppPass, cfg.SMTPHost)
	addr := cfg.SMTPHost + ":" + cfg.SMTPPort

	return smtp.SendMail(addr, auth, cfg.SMTPEmail, []string{toEmail}, msg)
}

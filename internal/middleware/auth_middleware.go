package middleware

import (
	"strings"

	"jaga-backend/internal/config"
	"jaga-backend/internal/repository"
	"jaga-backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequireAuth memvalidasi header "Authorization: Bearer <token>".
func RequireAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Token tidak ditemukan"})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseJWT(cfg.JWTSecret, tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Token tidak valid atau kedaluwarsa"})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

// RequireManageUsers hanya izinkan user dengan can_manage_users = true.
func RequireManageUsers(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Tidak terautentikasi"})
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "User tidak ditemukan"})
		}

		if !user.CanManageUsers {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Anda tidak punya izin untuk aksi ini"})
		}

		return c.Next()
	}
}

// RequireApproveReport hanya izinkan user dengan can_approve_report = true.
func RequireApproveReport(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Tidak terautentikasi"})
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "User tidak ditemukan"})
		}

		if !user.CanApproveReport {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Anda tidak punya izin untuk verifikasi laporan"})
		}

		return c.Next()
	}
}

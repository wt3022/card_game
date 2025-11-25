package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"card_game/internal/core/port"
)

// ========================================
// パスワードハッシュ実装
// ========================================

// passwordHasher パスワードのハッシュ化・検証を実装
type passwordHasher struct{}

// NewPasswordHasher 新しいパスワードハッシャーを作成
func NewPasswordHasher() port.PasswordHasher {
	return &passwordHasher{}
}

// HashPassword パスワードをハッシュ化
func (h *passwordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword パスワードを検証
func (h *passwordHasher) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"card_game/internal/core/port"
)

const (
	// デフォルトのbcryptコスト(推奨値: 12-14)
	defaultBcryptCost = 12
	minBcryptCost     = 10
	maxBcryptCost     = 14
	minPasswordLength = 8
	maxPasswordLength = 128
)

// ========================================
// パスワードハッシュ実装
// ========================================

// passwordHasher パスワードのハッシュ化・検証を実装
type passwordHasher struct {
	cost int
}

// NewPasswordHasher 新しいパスワードハッシャーを作成
func NewPasswordHasher() port.PasswordHasher {
	cost := defaultBcryptCost
	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		if c, err := strconv.Atoi(costStr); err == nil {
			if c >= minBcryptCost && c <= maxBcryptCost {
				cost = c
			}
		}
	}
	return &passwordHasher{cost: cost}
}

// HashPassword パスワードをハッシュ化
func (h *passwordHasher) HashPassword(password string) (string, error) {
	// パスワードのバリデーション
	if err := h.validatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword パスワードを検証
func (h *passwordHasher) VerifyPassword(hashedPassword, password string) error {
	// パスワードのバリデーション
	if err := h.validatePassword(password); err != nil {
		return err
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("パスワードが一致しません")
		}
		return fmt.Errorf("パスワードの検証に失敗しました: %w", err)
	}
	return nil
}

// validatePassword パスワードのバリデーション
func (h *passwordHasher) validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("パスワードは%d文字以上である必要があります", minPasswordLength)
	}
	if len(password) > maxPasswordLength {
		return fmt.Errorf("パスワードは%d文字以下である必要があります", maxPasswordLength)
	}
	return nil
}

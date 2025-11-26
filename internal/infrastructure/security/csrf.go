package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// ========================================
// CSRFトークン管理
// クロスサイトリクエストフォージェリ対策
// ========================================

// CSRFTokenManager CSRFトークンを管理
type CSRFTokenManager struct {
	tokens map[string]CSRFToken
	mu     sync.RWMutex
}

// CSRFToken CSRFトークン情報
type CSRFToken struct {
	Token     string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewCSRFTokenManager 新しいCSRFTokenManagerを作成
func NewCSRFTokenManager() *CSRFTokenManager {
	return &CSRFTokenManager{
		tokens: make(map[string]CSRFToken),
	}
}

// GenerateToken ユーザーのためのCSRFトークンを生成
func (m *CSRFTokenManager) GenerateToken(userID string) (string, error) {
	// 32バイトのランダムトークンを生成
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	// トークンを保存
	m.tokens[token] = CSRFToken{
		Token:     token,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour), // 1時間有効
	}

	return token, nil
}

// ValidateToken CSRFトークンを検証
func (m *CSRFTokenManager) ValidateToken(token string, userID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	csrfToken, exists := m.tokens[token]
	if !exists {
		return fmt.Errorf("CSRF token not found")
	}

	if csrfToken.UserID != userID {
		return fmt.Errorf("CSRF token user mismatch")
	}

	if time.Now().After(csrfToken.ExpiresAt) {
		return fmt.Errorf("CSRF token expired")
	}

	return nil
}

// InvalidateToken トークンを無効化
func (m *CSRFTokenManager) InvalidateToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tokens, token)
}

// Cleanup 期限切れトークンをクリーンアップ
func (m *CSRFTokenManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, csrfToken := range m.tokens {
		if now.After(csrfToken.ExpiresAt) {
			delete(m.tokens, token)
		}
	}
}

// StartCleanupRoutine クリーンアップルーチンを開始
func (m *CSRFTokenManager) StartCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			m.Cleanup()
		}
	}()
}

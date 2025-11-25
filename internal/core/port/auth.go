package port

import (
	"context"
	"time"
)

// ========================================
// 認証インターフェース
// ========================================

// JWTClaims JWTトークンのクレーム
type JWTClaims struct {
	UserID    string
	Username  string
	Role      string
	ExpiresAt time.Time
}

// TokenProvider JWTトークンの生成・検証を行うインターフェース
type TokenProvider interface {
	// GenerateToken ユーザー情報からJWTトークンを生成
	GenerateToken(userID, username, role string) (string, error)

	// ValidateToken トークンを検証してクレームを取得
	ValidateToken(tokenString string) (*JWTClaims, error)

	// ExtractTokenFromContext コンテキストからトークン情報を取得
	ExtractTokenFromContext(ctx context.Context) (*JWTClaims, error)
}

// PasswordHasher パスワードのハッシュ化・検証を行うインターフェース
type PasswordHasher interface {
	// HashPassword パスワードをハッシュ化
	HashPassword(password string) (string, error)

	// VerifyPassword パスワードを検証
	VerifyPassword(hashedPassword, password string) error
}

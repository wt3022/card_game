package jwt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"card_game/internal/core/port"
)

// ========================================
// JWT実装
// ========================================

// jwtProvider JWTトークンの生成・検証を実装
type jwtProvider struct {
	secret      []byte
	expiryHours int
}

// NewJWTProvider 新しいJWTプロバイダーを作成
func NewJWTProvider() (port.TokenProvider, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is not set")
	}

	expiryHours := 24 // デフォルト24時間
	if expiryStr := os.Getenv("JWT_EXPIRY_HOURS"); expiryStr != "" {
		if _, err := fmt.Sscanf(expiryStr, "%d", &expiryHours); err != nil {
			expiryHours = 24
		}
	}

	return &jwtProvider{
		secret:      []byte(secret),
		expiryHours: expiryHours,
	}, nil
}

// GenerateToken ユーザー情報からJWTトークンを生成
func (p *jwtProvider) GenerateToken(userID, username, role string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(p.expiryHours) * time.Hour)

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      expiresAt.Unix(),
		"iat":      now.Unix(),
		"iss":      "card_game",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken トークンを検証してクレームを取得
func (p *jwtProvider) ValidateToken(tokenString string) (*port.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 署名アルゴリズムを検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to extract claims")
	}

	// クレームを取得
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id claim is missing or invalid")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("username claim is missing or invalid")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("role claim is missing or invalid")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("exp claim is missing or invalid")
	}

	return &port.JWTClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: time.Unix(int64(exp), 0),
	}, nil
}

// ExtractTokenFromContext コンテキストからトークン情報を取得
func (p *jwtProvider) ExtractTokenFromContext(ctx context.Context) (*port.JWTClaims, error) {
	claims, ok := ctx.Value("jwt_claims").(*port.JWTClaims)
	if !ok || claims == nil {
		return nil, errors.New("jwt claims not found in context")
	}
	return claims, nil
}

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

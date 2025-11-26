package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"card_game/internal/core/port"
)

const (
	// デフォルトのトークン有効期限(時間)
	defaultExpiryHours = 24
	// 最小有効期限(1時間)
	minExpiryHours = 1
	// 最大有効期限(168時間 = 7日)
	maxExpiryHours = 168
	// JWTシークレットの最小長
	minSecretLength = 32
	// 発行者
	issuer = "card_game"
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
		return nil, errors.New("JWT_SECRET環境変数が設定されていません")
	}

	// シークレットの長さをチェック(セキュリティ向上)
	if len(secret) < minSecretLength {
		return nil, fmt.Errorf("JWT_SECRETは%d文字以上である必要があります", minSecretLength)
	}

	expiryHours := defaultExpiryHours
	if expiryStr := os.Getenv("JWT_EXPIRY_HOURS"); expiryStr != "" {
		if hours, err := strconv.Atoi(expiryStr); err == nil {
			if hours >= minExpiryHours && hours <= maxExpiryHours {
				expiryHours = hours
			}
		}
	}

	return &jwtProvider{
		secret:      []byte(secret),
		expiryHours: expiryHours,
	}, nil
}

// GenerateToken ユーザー情報からJWTトークンを生成
func (p *jwtProvider) GenerateToken(userID, username, role string) (string, error) {
	// 入力バリデーション
	if err := p.validateTokenInput(userID, username, role); err != nil {
		return "", err
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(p.expiryHours) * time.Hour)

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      expiresAt.Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Unix(), // Not Before: 即座に有効
		"iss":      issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("トークンの署名に失敗しました: %w", err)
	}

	return tokenString, nil
}

// ValidateToken トークンを検証してクレームを取得
func (p *jwtProvider) ValidateToken(tokenString string) (*port.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 署名アルゴリズムを検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("予期しない署名メソッド: %v", token.Header["alg"])
		}
		return p.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("トークンの解析に失敗しました: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("トークンが無効です")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("クレームの抽出に失敗しました")
	}

	// クレームを取得
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_idクレームが欠落しているか無効です")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("usernameクレームが欠落しているか無効です")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("roleクレームが欠落しているか無効です")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("expクレームが欠落しているか無効です")
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
		return nil, errors.New("コンテキストにjwtクレームが見つかりません")
	}
	return claims, nil
}

// validateTokenInput トークン生成時の入力をバリデーション
func (p *jwtProvider) validateTokenInput(userID, username, role string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("userIDは必須です")
	}
	if strings.TrimSpace(username) == "" {
		return errors.New("usernameは必須です")
	}
	if strings.TrimSpace(role) == "" {
		return errors.New("roleは必須です")
	}
	return nil
}

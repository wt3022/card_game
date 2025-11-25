package service

import (
	"errors"
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/repository"
)

// ========================================
// 認証サービス
// ========================================

// AuthService 認証ビジネスロジック
type AuthService struct {
	userRepo       repository.UserRepository
	tokenProvider  port.TokenProvider
	passwordHasher port.PasswordHasher
	logger         port.Logger
}

// NewAuthService 新しい認証サービスを作成
func NewAuthService(
	userRepo repository.UserRepository,
	tokenProvider port.TokenProvider,
	passwordHasher port.PasswordHasher,
	logger port.Logger,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenProvider:  tokenProvider,
		passwordHasher: passwordHasher,
		logger:         logger,
	}
}

// Login ログイン処理
func (s *AuthService) Login(username, password string) (*entity.LoginResponse, error) {
	// ユーザーを検索
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// パスワードを検証
	if err := s.passwordHasher.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// JWTトークンを生成
	token, err := s.tokenProvider.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &entity.LoginResponse{
		AccessToken: token,
		UserID:      user.ID,
		Username:    user.Username,
		Role:        string(user.Role),
	}, nil
}

// ValidateToken トークンを検証
func (s *AuthService) ValidateToken(tokenString string) (*port.JWTClaims, error) {
	return s.tokenProvider.ValidateToken(tokenString)
}

// CreateUser ユーザーを作成（管理者用）
func (s *AuthService) CreateUser(username, password string, role entity.UserRole) (*entity.User, error) {
	// 既存ユーザーをチェック
	_, err := s.userRepo.FindByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// パスワードをハッシュ化
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// ユーザーIDを生成（簡易版、実際はUUID等を使用）
	userID := fmt.Sprintf("user_%s", username)

	user := &entity.User{
		ID:           userID,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

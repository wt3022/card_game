package service

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// 認証サービス
// ========================================

// AuthService 認証ビジネスロジック
type AuthService struct {
	userRepo       port.UserRepository
	tokenProvider  port.TokenProvider
	passwordHasher port.PasswordHasher
	logger         port.Logger
}

// NewAuthService 新しい認証サービスを作成
func NewAuthService(
	userRepo port.UserRepository,
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
	// 入力バリデーション
	if err := entity.ValidateUsername(username); err != nil {
		s.logger.Error("Username validation failed for '%s': %v", username, err)
		return nil, fmt.Errorf("認証情報が無効です")
	}

	// ユーザーを検索
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		s.logger.Error("User not found: '%s'", username)
		return nil, fmt.Errorf("認証情報が無効です")
	}

	// パスワードを検証（検証はpasswordHasher内で行われる）
	if err := s.passwordHasher.VerifyPassword(user.PasswordHash, password); err != nil {
		s.logger.Error("Password verification failed for user '%s': %v", username, err)
		return nil, fmt.Errorf("認証情報が無効です")
	}

	// JWTトークンを生成
	token, err := s.tokenProvider.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("トークンの生成に失敗しました: %w", err)
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
	// 入力バリデーション
	if err := entity.ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := entity.ValidateUserRole(role); err != nil {
		return nil, err
	}

	// 既存ユーザーをチェック
	_, err := s.userRepo.FindByUsername(username)
	if err == nil {
		return nil, entity.NewErrAlreadyExists("user", username)
	}

	// パスワードをハッシュ化（検証はpasswordHasher内で行われる）
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
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
		return nil, fmt.Errorf("ユーザーの作成に失敗しました: %w", err)
	}

	return user, nil
}

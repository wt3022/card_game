package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/repository"
)

// ========================================
// 初期ユーザー作成
// ========================================

// InitializeDefaultAdmin デフォルトの管理者ユーザーを初期化
// 既に管理者ユーザーが存在する場合は何もしない
func InitializeDefaultAdmin(
	userRepo repository.UserRepository,
	passwordHasher port.PasswordHasher,
	logger port.Logger,
) error {
	// 管理者ユーザーが既に存在するかチェック
	// 任意の管理者ユーザーを探す（簡易実装）
	adminUsername := getEnv("ADMIN_USERNAME", "admin")
	_, err := userRepo.FindByUsername(adminUsername)
	if err == nil {
		// 既に存在する
		logger.Info("Admin user already exists: %s", adminUsername)
		return nil
	}

	// エラーメッセージを確認して、ユーザーが見つからない場合は正常な動作として扱う
	if err != nil && !strings.Contains(err.Error(), "user not found") {
		// 予期しないエラーの場合のみログを出力
		logger.Info("⚠️  Error checking admin user: %v", err)
	}

	// 管理者ユーザーが存在しない場合、作成
	adminPassword := getEnv("ADMIN_PASSWORD", "admin123")
	if adminPassword == "admin123" {
		logger.Info("⚠️  Using default admin password. Please change it in production!")
	}

	// パスワードをハッシュ化
	passwordHash, err := passwordHasher.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// ユーザーIDを生成（UUID）
	userID := uuid.New().String()

	user := &entity.User{
		ID:           userID,
		Username:     adminUsername,
		PasswordHash: passwordHash,
		Role:         entity.UserRoleAdmin,
		CreatedAt:    time.Now(), // 追加
		UpdatedAt:    time.Now(), // 追加
	}

	if err := userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logger.Info("✅ Created default admin user: %s (ID: %s)", adminUsername, userID)
	return nil
}

// getEnv 環境変数を取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

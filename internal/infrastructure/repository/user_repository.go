package repository

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence/model"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ========================================
// ユーザーリポジトリ実装
// ========================================

// userRepository ユーザーリポジトリの実装
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 新しいユーザーリポジトリを作成
func NewUserRepository(db *gorm.DB) port.UserRepository {
	return &userRepository{db: db}
}

// toEntityUser UserModelをentity.Userに変換
func toEntityUser(userModel *model.UserModel) *entity.User {
	return &entity.User{
		ID:           userModel.ID,
		Username:     userModel.Username,
		PasswordHash: userModel.PasswordHash,
		Role:         entity.UserRole(userModel.Role),
		CreatedAt:    userModel.CreatedAt,
		UpdatedAt:    userModel.UpdatedAt,
	}
}

// toGormUser entity.UserをUserModelに変換
func toGormUser(user *entity.User) *model.UserModel {
	userModel := &model.UserModel{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
	}
	// CreatedAtとUpdatedAtがゼロ値でない場合のみ設定
	if !user.CreatedAt.IsZero() {
		userModel.CreatedAt = user.CreatedAt
	}
	if !user.UpdatedAt.IsZero() {
		userModel.UpdatedAt = user.UpdatedAt
	}
	return userModel
}

// Create ユーザーを作成
func (r *userRepository) Create(user *entity.User) error {
	userModel := toGormUser(user)
	if err := r.db.Create(userModel).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByID IDでユーザーを検索
func (r *userRepository) FindByID(id string) (*entity.User, error) {
	var userModel model.UserModel
	if err := r.db.First(&userModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return toEntityUser(&userModel), nil
}

// FindByUsername ユーザー名でユーザーを検索
func (r *userRepository) FindByUsername(username string) (*entity.User, error) {
	var userModel model.UserModel
	// ErrRecordNotFoundの場合はログを抑制するためにSessionを使用
	err := r.db.Session(&gorm.Session{
		Logger: r.db.Logger.LogMode(logger.Silent),
	}).Where("username = ?", username).First(&userModel).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	return toEntityUser(&userModel), nil
}

// Update ユーザーを更新
func (r *userRepository) Update(user *entity.User) error {
	userModel := toGormUser(user)
	if err := r.db.Save(userModel).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete ユーザーを削除
func (r *userRepository) Delete(id string) error {
	if err := r.db.Delete(&model.UserModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

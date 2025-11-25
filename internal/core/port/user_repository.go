package port

import (
	"card_game/internal/core/entity"
)

// ========================================
// ユーザーリポジトリインターフェース
// ========================================

// UserRepository ユーザーリポジトリインターフェース
type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id string) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	Update(user *entity.User) error
	Delete(id string) error
}

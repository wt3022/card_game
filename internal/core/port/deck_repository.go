package port

import (
	"card_game/internal/core/entity"
	"context"
)

// DeckRepository はデッキの永続化を担当するポート
type DeckRepository interface {
	// Create は新しいデッキを作成
	Create(ctx context.Context, deck *entity.Deck) error

	// FindByID はIDでデッキを取得
	FindByID(ctx context.Context, id string) (*entity.Deck, error)

	// FindByUserID はユーザーIDでデッキ一覧を取得
	FindByUserID(ctx context.Context, userID string) ([]*entity.Deck, error)

	// FindAll はすべてのデッキを取得（管理者用）
	FindAll(ctx context.Context) ([]*entity.Deck, error)

	// Update は既存のデッキを更新
	Update(ctx context.Context, deck *entity.Deck) error

	// Delete はデッキを削除
	Delete(ctx context.Context, id string) error

	// ExistsByID はIDでデッキの存在を確認
	ExistsByID(ctx context.Context, id string) (bool, error)
}

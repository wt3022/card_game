package port

import (
	"card_game/internal/core/entity"
)

// CardRepository カードリポジトリインターフェース
type CardRepository interface {
	Create(card *entity.Card) error
	FindByID(id string) (*entity.Card, error)
	FindAll() ([]*entity.Card, error)
	FindByType(cardType entity.CardType) ([]*entity.Card, error)
	Update(card *entity.Card) error
	Delete(id string) error
	SaveCardEffect(cardID string, cardEffectModel interface{}) error
	// GetCardEffectAsProto CardEffectをProto形式で取得（adapterレイヤー用）
	GetCardEffectAsProto(cardID string) (interface{}, error)
}

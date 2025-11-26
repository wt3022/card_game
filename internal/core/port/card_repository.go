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
	SaveCardEffect(cardID string, cardEffect *entity.CardEffect) error
	GetCardEffect(cardID string) (*entity.CardEffect, error)
	GenerateEffectDescription(cardID string) (string, error)
}

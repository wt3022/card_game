package port

import (
	"card_game/internal/core/entity"
)

// CardListResult ページネーション結果
type CardListResult struct {
	Cards      []*entity.Card
	TotalCount int
	Page       int
	PageSize   int
}

// CardRepository カードリポジトリインターフェース
type CardRepository interface {
	Create(card *entity.Card) error
	FindByID(id string) (*entity.Card, error)
	FindAll() ([]*entity.Card, error)
	FindAllWithPagination(page, pageSize int) (*CardListResult, error)
	FindByType(cardType entity.CardType) ([]*entity.Card, error)
	FindByTypeWithPagination(cardType entity.CardType, page, pageSize int) (*CardListResult, error)
	Update(card *entity.Card) error
	Delete(id string) error
	SaveCardEffect(cardID string, cardEffect *entity.CardEffect) error
	GetCardEffect(cardID string) (*entity.CardEffect, error)
	GenerateEffectDescription(cardID string) (string, error)
}

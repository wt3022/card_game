package service

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence/model"
	"card_game/internal/infrastructure/repository"
)

// ========================================
// カード管理サービス
// ========================================

// CardService カード管理ビジネスロジック
type CardService struct {
	cardRepo repository.CardRepository
	logger   port.Logger
}

// NewCardService 新しいカード管理サービスを作成
func NewCardService(
	cardRepo repository.CardRepository,
	logger port.Logger,
) *CardService {
	return &CardService{
		cardRepo: cardRepo,
		logger:   logger,
	}
}

// CreateCard カードを作成
func (s *CardService) CreateCard(card *entity.Card) error {
	// バリデーション
	if err := s.validateCard(card); err != nil {
		return err
	}

	// 既存のカードをチェック
	_, err := s.cardRepo.FindByID(card.ID)
	if err == nil {
		return fmt.Errorf("card with id %s already exists", card.ID)
	}

	// カードを作成
	if err := s.cardRepo.Create(card); err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	s.logger.Info("Card created: %s (%s)", card.ID, card.Name)
	return nil
}

// GetCard IDでカードを取得
func (s *CardService) GetCard(id string) (*entity.Card, error) {
	card, err := s.cardRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get card: %w", err)
	}
	return card, nil
}

// ListCards すべてのカードを取得
func (s *CardService) ListCards() ([]*entity.Card, error) {
	cards, err := s.cardRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}
	return cards, nil
}

// ListCardsByType タイプでカードを取得
func (s *CardService) ListCardsByType(cardType entity.CardType) ([]*entity.Card, error) {
	cards, err := s.cardRepo.FindByType(cardType)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards by type: %w", err)
	}
	return cards, nil
}

// UpdateCard カードを更新
func (s *CardService) UpdateCard(card *entity.Card) error {
	// バリデーション
	if err := s.validateCard(card); err != nil {
		return err
	}

	// 既存のカードをチェック
	_, err := s.cardRepo.FindByID(card.ID)
	if err != nil {
		return fmt.Errorf("card with id %s not found", card.ID)
	}

	// カードを更新
	if err := s.cardRepo.Update(card); err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	s.logger.Info("Card updated: %s (%s)", card.ID, card.Name)
	return nil
}

// DeleteCard カードを削除
func (s *CardService) DeleteCard(id string) error {
	// 既存のカードをチェック
	card, err := s.cardRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("card with id %s not found", id)
	}

	// カードを削除
	if err := s.cardRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	s.logger.Info("Card deleted: %s (%s)", card.ID, card.Name)
	return nil
}

// SaveCardEffect カード効果を保存（model構造を直接受け取る）
func (s *CardService) SaveCardEffect(cardID string, cardEffectModel *model.CardEffectModel) error {
	if err := s.cardRepo.SaveCardEffect(cardID, cardEffectModel); err != nil {
		return fmt.Errorf("failed to save card effect: %w", err)
	}
	s.logger.Info("Card effect saved for card: %s", cardID)
	return nil
}

// validateCard カードのバリデーション
func (s *CardService) validateCard(card *entity.Card) error {
	if card.ID == "" {
		return fmt.Errorf("card id is required")
	}
	if card.Name == "" {
		return fmt.Errorf("card name is required")
	}
	if card.Type == "" {
		return fmt.Errorf("card type is required")
	}
	if card.Cost < 0 {
		return fmt.Errorf("card cost must be non-negative")
	}

	// ユニットカードの場合、攻撃力と防御力が必要
	if card.Type == entity.CardTypeUnit {
		if card.Attack == nil || card.Defense == nil {
			return fmt.Errorf("unit card must have attack and defense")
		}
		if *card.Attack < 0 || *card.Defense < 0 {
			return fmt.Errorf("attack and defense must be non-negative")
		}
	}

	// スペルまたはリーダーカードの場合、特性は設定できない
	if card.Type == entity.CardTypeSpell || card.Type == entity.CardTypeLeader {
		if len(card.Traits) > 0 {
			return fmt.Errorf("spell and leader cards cannot have traits")
		}
	}

	// スペルまたはリーダーカードの場合、攻撃力と防御力は設定できない
	if card.Type == entity.CardTypeSpell || card.Type == entity.CardTypeLeader {
		if card.Attack != nil || card.Defense != nil {
			return fmt.Errorf("spell and leader cards cannot have attack or defense")
		}
	}

	return nil
}

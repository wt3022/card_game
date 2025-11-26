package service

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// カード管理サービス
// ========================================

// CardService カード管理ビジネスロジック
type CardService struct {
	cardRepo port.CardRepository
	logger   port.Logger
}

// NewCardService 新しいカード管理サービスを作成
func NewCardService(
	cardRepo port.CardRepository,
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
	if err := card.Validate(); err != nil {
		s.logger.Error("Card validation failed: %v", err)
		return err
	}

	// IDの重複チェック（IDが明示的に指定されている場合）
	if card.ID != "" {
		_, err := s.cardRepo.FindByID(card.ID)
		if err == nil {
			return entity.NewErrAlreadyExists("card", card.ID)
		}
	}

	// カードを作成
	if err := s.cardRepo.Create(card); err != nil {
		s.logger.Error("Failed to create card: %v", err)
		return fmt.Errorf("カードの作成に失敗しました: %w", err)
	}

	s.logger.Info("Card created: %s (%s)", card.ID, card.Name)
	return nil
}

// GetCard IDでカードを取得
func (s *CardService) GetCard(id string) (*entity.Card, error) {
	card, err := s.cardRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("カードの取得に失敗しました: %w", err)
	}
	return card, nil
}

// ListCards すべてのカードを取得
func (s *CardService) ListCards() ([]*entity.Card, error) {
	cards, err := s.cardRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("カード一覧の取得に失敗しました: %w", err)
	}
	return cards, nil
}

// ListCardsByType タイプでカードを取得
func (s *CardService) ListCardsByType(cardType entity.CardType) ([]*entity.Card, error) {
	cards, err := s.cardRepo.FindByType(cardType)
	if err != nil {
		return nil, fmt.Errorf("タイプ別カード一覧の取得に失敗しました: %w", err)
	}
	return cards, nil
}

// UpdateCard カードを更新
func (s *CardService) UpdateCard(card *entity.Card) error {
	// バリデーション
	if err := card.Validate(); err != nil {
		s.logger.Error("Card validation failed: %v", err)
		return err
	}

	// 既存のカードをチェック
	_, err := s.cardRepo.FindByID(card.ID)
	if err != nil {
		return entity.NewErrNotFound("card", card.ID)
	}

	// カードを更新
	if err := s.cardRepo.Update(card); err != nil {
		s.logger.Error("Failed to update card: %v", err)
		return fmt.Errorf("カードの更新に失敗しました: %w", err)
	}

	s.logger.Info("Card updated: %s (%s)", card.ID, card.Name)
	return nil
}

// DeleteCard カードを削除
func (s *CardService) DeleteCard(id string) error {
	// 入力バリデーション
	if id == "" {
		return entity.NewErrInvalidInput("card.id", "カードIDは必須です")
	}

	// 既存のカードをチェック
	card, err := s.cardRepo.FindByID(id)
	if err != nil {
		return entity.NewErrNotFound("card", id)
	}

	// カードを削除
	if err := s.cardRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete card: %v", err)
		return fmt.Errorf("カードの削除に失敗しました: %w", err)
	}

	s.logger.Info("Card deleted: %s (%s)", card.ID, card.Name)
	return nil
}

// SaveCardEffect カード効果を保存
func (s *CardService) SaveCardEffect(cardID string, cardEffect *entity.CardEffect) error {
	if err := s.cardRepo.SaveCardEffect(cardID, cardEffect); err != nil {
		return fmt.Errorf("カード効果の保存に失敗しました: %w", err)
	}
	s.logger.Info("Card effect saved for card: %s", cardID)
	return nil
}

// GetCardEffect カード効果を取得
func (s *CardService) GetCardEffect(cardID string) (*entity.CardEffect, error) {
	cardEffect, err := s.cardRepo.GetCardEffect(cardID)
	if err != nil {
		return nil, fmt.Errorf("カード効果の取得に失敗しました: %w", err)
	}
	return cardEffect, nil
}

// GenerateEffectDescription カード効果から効果テキストを生成
func (s *CardService) GenerateEffectDescription(cardID string) (string, error) {
	description, err := s.cardRepo.GenerateEffectDescription(cardID)
	if err != nil {
		return "", fmt.Errorf("効果テキストの生成に失敗しました: %w", err)
	}
	return description, nil
}

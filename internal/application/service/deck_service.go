package service

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"context"

	"github.com/google/uuid"
)

// DeckService はデッキ管理のアプリケーションサービス
type DeckService struct {
	deckRepo port.DeckRepository
	cardRepo port.CardRepository
	logger   port.Logger
}

// NewDeckService は新しいDeckServiceを作成
func NewDeckService(
	deckRepo port.DeckRepository,
	cardRepo port.CardRepository,
	logger port.Logger,
) *DeckService {
	return &DeckService{
		deckRepo: deckRepo,
		cardRepo: cardRepo,
		logger:   logger,
	}
}

// CreateDeck は新しいデッキを作成
func (s *DeckService) CreateDeck(ctx context.Context, name, description, userID string, cardIDs []string) (*entity.Deck, error) {
	// カードの存在確認
	if err := s.validateCardIDs(ctx, cardIDs); err != nil {
		return nil, err
	}

	// デッキエンティティの作成
	deckID := uuid.New().String()
	deck, err := entity.NewDeck(deckID, name, description, userID, cardIDs)
	if err != nil {
		return nil, err
	}

	// 永続化
	if err := s.deckRepo.Create(ctx, deck); err != nil {
		s.logger.Error("Failed to create deck", "error", err, "deck_id", deckID)
		return nil, err
	}

	s.logger.Info("Deck created", "deck_id", deckID, "user_id", userID)
	return deck, nil
}

// GetDeck はIDでデッキを取得
func (s *DeckService) GetDeck(ctx context.Context, id string) (*entity.Deck, error) {
	// 入力バリデーション
	if id == "" {
		return nil, entity.NewErrInvalidInput("deck.id", "デッキIDは必須です")
	}

	deck, err := s.deckRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get deck", "error", err, "deck_id", id)
		return nil, entity.NewErrNotFound("deck", id)
	}
	return deck, nil
}

// ListDecksByUser はユーザーIDでデッキ一覧を取得
func (s *DeckService) ListDecksByUser(ctx context.Context, userID string) ([]*entity.Deck, error) {
	decks, err := s.deckRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to list decks by user", "error", err, "user_id", userID)
		return nil, err
	}
	return decks, nil
}

// ListAllDecks はすべてのデッキを取得（管理者用）
func (s *DeckService) ListAllDecks(ctx context.Context) ([]*entity.Deck, error) {
	decks, err := s.deckRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Failed to list all decks", "error", err)
		return nil, err
	}
	return decks, nil
}

// UpdateDeck はデッキを更新
func (s *DeckService) UpdateDeck(ctx context.Context, id, name, description string, cardIDs []string) (*entity.Deck, error) {
	// 既存のデッキを取得
	deck, err := s.deckRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// カードの存在確認
	if err := s.validateCardIDs(ctx, cardIDs); err != nil {
		return nil, err
	}

	// デッキ情報を更新
	if err := deck.UpdateInfo(name, description); err != nil {
		return nil, err
	}

	// カードリストを更新
	if err := deck.UpdateCards(cardIDs); err != nil {
		return nil, err
	}

	// 永続化
	if err := s.deckRepo.Update(ctx, deck); err != nil {
		s.logger.Error("Failed to update deck", "error", err, "deck_id", id)
		return nil, err
	}

	s.logger.Info("Deck updated", "deck_id", id)
	return deck, nil
}

// DeleteDeck はデッキを削除
func (s *DeckService) DeleteDeck(ctx context.Context, id string) error {
	// 入力バリデーション
	if id == "" {
		return entity.NewErrInvalidInput("deck.id", "デッキIDは必須です")
	}

	if err := s.deckRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete deck", "error", err, "deck_id", id)
		return err
	}

	s.logger.Info("Deck deleted", "deck_id", id)
	return nil
}

// validateCardIDs はカードIDの存在を確認
func (s *DeckService) validateCardIDs(ctx context.Context, cardIDs []string) error {
	// 空のカードリストをチェック
	if len(cardIDs) == 0 {
		return entity.NewErrInvalidDeck("cards", "カードリストは空にできません")
	}

	// 重複を除いてユニークなカードIDを取得
	uniqueCardIDs := make(map[string]bool)
	for _, cardID := range cardIDs {
		// 空のカードIDをチェック
		if cardID == "" {
			return entity.NewErrInvalidDeck("cards", "無効なカードIDが含まれています")
		}
		uniqueCardIDs[cardID] = true
	}

	// 各カードの存在確認
	for cardID := range uniqueCardIDs {
		_, err := s.cardRepo.FindByID(cardID)
		if err != nil {
			return entity.NewErrInvalidDeck("cards", "存在しないカードが含まれています: "+cardID)
		}
	}

	return nil
}

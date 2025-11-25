package repository

import (
	"card_game/internal/core/entity"
	"card_game/internal/infrastructure/persistence/model"
	"context"

	"gorm.io/gorm"
)

type deckRepository struct {
	db *gorm.DB
}

// NewDeckRepository はDeckRepositoryの実装を作成
func NewDeckRepository(db *gorm.DB) *deckRepository {
	return &deckRepository{db: db}
}

// Create は新しいデッキを作成
func (r *deckRepository) Create(ctx context.Context, deck *entity.Deck) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deckModel := toDeckModelWithoutCards(deck)
		if err := tx.Create(deckModel).Error; err != nil {
			return err
		}

		// カードの関連を作成
		deckCards := make([]model.DeckCardModel, len(deck.CardIDs))
		for i, cardID := range deck.CardIDs {
			deckCards[i] = model.DeckCardModel{
				DeckID:   deck.ID,
				CardID:   cardID,
				Position: i,
			}
		}

		if len(deckCards) > 0 {
			if err := tx.Create(&deckCards).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindByID はIDでデッキを取得
func (r *deckRepository) FindByID(ctx context.Context, id string) (*entity.Deck, error) {
	var deckModel model.DeckModel
	if err := r.db.WithContext(ctx).
		Preload("Cards", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Where("id = ?", id).
		First(&deckModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.NewErrNotFound("Deck", id)
		}
		return nil, err
	}

	return toDeckEntity(&deckModel), nil
}

// FindByUserID はユーザーIDでデッキ一覧を取得
func (r *deckRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Deck, error) {
	var deckModels []model.DeckModel
	if err := r.db.WithContext(ctx).
		Preload("Cards", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Where("user_id = ?", userID).
		Find(&deckModels).Error; err != nil {
		return nil, err
	}

	decks := make([]*entity.Deck, 0, len(deckModels))
	for i := range deckModels {
		deck := toDeckEntity(&deckModels[i])
		decks = append(decks, deck)
	}

	return decks, nil
}

// FindAll はすべてのデッキを取得
func (r *deckRepository) FindAll(ctx context.Context) ([]*entity.Deck, error) {
	var deckModels []model.DeckModel
	if err := r.db.WithContext(ctx).
		Preload("Cards", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Find(&deckModels).Error; err != nil {
		return nil, err
	}

	decks := make([]*entity.Deck, 0, len(deckModels))
	for i := range deckModels {
		deck := toDeckEntity(&deckModels[i])
		decks = append(decks, deck)
	}

	return decks, nil
}

// Update は既存のデッキを更新
func (r *deckRepository) Update(ctx context.Context, deck *entity.Deck) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// デッキ情報を更新
		deckModel := toDeckModelWithoutCards(deck)
		if err := tx.Save(deckModel).Error; err != nil {
			return err
		}

		// 既存のカード関連を削除
		if err := tx.Where("deck_id = ?", deck.ID).Delete(&model.DeckCardModel{}).Error; err != nil {
			return err
		}

		// 新しいカード関連を作成
		deckCards := make([]model.DeckCardModel, len(deck.CardIDs))
		for i, cardID := range deck.CardIDs {
			deckCards[i] = model.DeckCardModel{
				DeckID:   deck.ID,
				CardID:   cardID,
				Position: i,
			}
		}

		if len(deckCards) > 0 {
			if err := tx.Create(&deckCards).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete はデッキを削除
func (r *deckRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.DeckModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.NewErrNotFound("Deck", id)
	}
	return nil
}

// ExistsByID はIDでデッキの存在を確認
func (r *deckRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DeckModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// toDeckModelWithoutCards はエンティティをモデルに変換（カード関連なし）
func toDeckModelWithoutCards(deck *entity.Deck) *model.DeckModel {
	return &model.DeckModel{
		ID:          deck.ID,
		Name:        deck.Name,
		Description: deck.Description,
		UserID:      deck.UserID,
		CreatedAt:   deck.CreatedAt,
		UpdatedAt:   deck.UpdatedAt,
	}
}

// toDeckEntity はモデルをエンティティに変換
func toDeckEntity(deckModel *model.DeckModel) *entity.Deck {
	cardIDs := make([]string, len(deckModel.Cards))
	for i, card := range deckModel.Cards {
		cardIDs[i] = card.CardID
	}

	return &entity.Deck{
		ID:          deckModel.ID,
		Name:        deckModel.Name,
		Description: deckModel.Description,
		CardIDs:     cardIDs,
		UserID:      deckModel.UserID,
		CreatedAt:   deckModel.CreatedAt,
		UpdatedAt:   deckModel.UpdatedAt,
	}
}

package entity

import (
	"fmt"
	"time"
)

const (
	// DeckSize デッキの枚数（固定）
	DeckSize = 40
	// MaxCardCopies 同じカードの最大枚数
	MaxCardCopies = 3
	// MaxDeckNameLength デッキ名の最大長
	MaxDeckNameLength = 100
	// MaxDeckDescriptionLength デッキ説明の最大長
	MaxDeckDescriptionLength = 500
)

// Deck はデッキを表すドメインエンティティ
type Deck struct {
	ID          string
	Name        string
	Description string
	CardIDs     []string
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDeck は新しいDeckエンティティを作成
func NewDeck(id, name, description, userID string, cardIDs []string) (*Deck, error) {
	now := time.Now()
	deck := &Deck{
		ID:          id,
		Name:        name,
		Description: description,
		CardIDs:     cardIDs,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := deck.Validate(); err != nil {
		return nil, err
	}

	return deck, nil
}

// Validate はデッキの妥当性を検証
func (d *Deck) Validate() error {
	// 名前検証
	if d.Name == "" {
		return NewErrInvalidDeck("name", "デッキ名は必須です")
	}
	if len(d.Name) > MaxDeckNameLength {
		return NewErrInvalidDeck("name", fmt.Sprintf("デッキ名は%d文字以内である必要があります", MaxDeckNameLength))
	}

	// 説明検証
	if len(d.Description) > MaxDeckDescriptionLength {
		return NewErrInvalidDeck("description", fmt.Sprintf("デッキ説明は%d文字以内である必要があります", MaxDeckDescriptionLength))
	}

	// ユーザーID検証
	if d.UserID == "" {
		return NewErrInvalidDeck("user_id", "ユーザーIDは必須です")
	}

	// カード数検証
	if len(d.CardIDs) == 0 {
		return NewErrInvalidDeck("cards", "デッキにカードが含まれていません")
	}

	// デッキの枚数は40枚固定
	if len(d.CardIDs) != DeckSize {
		return NewErrInvalidDeck("cards", fmt.Sprintf("デッキはちょうど%d枚である必要があります", DeckSize))
	}

	// 同じカードの枚数チェック
	cardCount := make(map[string]int)
	for _, cardID := range d.CardIDs {
		// 空のカードIDをチェック
		if cardID == "" {
			return NewErrInvalidDeck("cards", "無効なカードIDが含まれています")
		}

		cardCount[cardID]++
		if cardCount[cardID] > MaxCardCopies {
			return NewErrInvalidDeck("cards", fmt.Sprintf("同じカードは%d枚までしか入れられません", MaxCardCopies))
		}
	}

	return nil
}

// UpdateCards はカードリストを更新
func (d *Deck) UpdateCards(cardIDs []string) error {
	oldCardIDs := d.CardIDs
	d.CardIDs = cardIDs
	d.UpdatedAt = time.Now()

	if err := d.Validate(); err != nil {
		d.CardIDs = oldCardIDs // ロールバック
		return err
	}

	return nil
}

// UpdateInfo はデッキ名と説明を更新
func (d *Deck) UpdateInfo(name, description string) error {
	if name == "" {
		return NewErrInvalidDeck("name", "デッキ名は必須です")
	}
	if len(name) > MaxDeckNameLength {
		return NewErrInvalidDeck("name", fmt.Sprintf("デッキ名は%d文字以内である必要があります", MaxDeckNameLength))
	}
	if len(description) > MaxDeckDescriptionLength {
		return NewErrInvalidDeck("description", fmt.Sprintf("デッキ説明は%d文字以内である必要があります", MaxDeckDescriptionLength))
	}

	d.Name = name
	d.Description = description
	d.UpdatedAt = time.Now()
	return nil
}

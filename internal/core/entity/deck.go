package entity

import "time"

const (
	// DeckSize デッキの固定枚数
	DeckSize = 40
	// MaxCardCopies 同じカードの最大枚数
	MaxCardCopies = 3
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
	if d.Name == "" {
		return NewErrInvalidDeck("name", "デッキ名は必須です")
	}

	if len(d.CardIDs) == 0 {
		return NewErrInvalidDeck("cards", "デッキにカードが含まれていません")
	}

	// デッキの枚数は厳密に40枚
	if len(d.CardIDs) != DeckSize {
		return NewErrInvalidDeck("cards", "デッキは正確に40枚である必要があります")
	}

	// 同じカードの枚数チェック
	cardCount := make(map[string]int)
	for _, cardID := range d.CardIDs {
		cardCount[cardID]++
		if cardCount[cardID] > MaxCardCopies {
			return NewErrInvalidDeck("cards", "同じカードは3枚までしか入れられません")
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

	d.Name = name
	d.Description = description
	d.UpdatedAt = time.Now()
	return nil
}

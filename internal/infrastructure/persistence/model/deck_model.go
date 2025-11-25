package model

import "time"

// DeckModel はデッキのデータベースモデル
type DeckModel struct {
	ID          string          `gorm:"primaryKey;type:varchar(255)"`
	Name        string          `gorm:"type:varchar(255);not null"`
	Description string          `gorm:"type:text"`
	UserID      string          `gorm:"type:varchar(255);not null;index"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"`
	Cards       []DeckCardModel `gorm:"foreignKey:DeckID;constraint:OnDelete:CASCADE"`
}

// TableName はテーブル名を指定
func (DeckModel) TableName() string {
	return "decks"
}

// DeckCardModel はデッキとカードの関連を表すモデル
type DeckCardModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	DeckID    string    `gorm:"type:varchar(255);not null;index:idx_deck_card"`
	CardID    string    `gorm:"type:varchar(255);not null;index:idx_deck_card"`
	Position  int       `gorm:"not null"` // デッキ内の位置（0-39）
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName はテーブル名を指定
func (DeckCardModel) TableName() string {
	return "deck_cards"
}

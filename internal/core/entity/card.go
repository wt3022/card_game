package entity

// ========================================
// カード
// カードの基本構造を表す構造体を定義する
// ========================================

// カードの種類
type CardType string

const (
	CardTypeUnit   CardType = "Unit"   // ユニットカード
	CardTypeSpell  CardType = "Spell"  // スペルカード
	CardTypeLeader CardType = "Leader" // リーダーカード
)

// カードの基本構造
type Card struct {
	ID         string      `json:"id"`                    // カードID
	Name       string      `json:"name"`                  // カード名
	Type       CardType    `json:"type"`                  // カード種別
	Cost       int         `json:"cost"`                  // マナコスト
	Attack     *int        `json:"attack,omitempty"`      // 攻撃力
	Defense    *int        `json:"defense,omitempty"`     // 守備力
	Effect     string      `json:"effect"`                // 表示用の効果テキスト
	CardEffect *CardEffect `json:"card_effect,omitempty"` // 実際の効果定義
	Traits     []Trait     `json:"traits"`                // カードが持つ特性
}

// カードが指定された特性を持っているか確認
func (c *Card) HasTrait(trait Trait) bool {
	for _, t := range c.Traits {
		if t == trait {
			return true
		}
	}
	return false
}

// カードがユニットか確認
func (c *Card) IsUnit() bool {
	return c.Type == CardTypeUnit
}

// カードがスペルか確認
func (c *Card) IsSpell() bool {
	return c.Type == CardTypeSpell
}

// カードがリーダーか確認
func (c *Card) IsLeader() bool {
	return c.Type == CardTypeLeader
}

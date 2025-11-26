package entity

import "fmt"

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

const (
	// カードバリデーション定数
	MaxCardNameLength   = 50
	MaxCardEffectLength = 500
	MaxCardCost         = 10
	MaxCardAttack       = 20
	MaxCardDefense      = 20
	MaxCardTraits       = 5
)

// Validate カードのバリデーション
func (c *Card) Validate() error {
	// ID検証
	if c.ID == "" {
		return NewErrInvalidInput("card.id", "カードIDは必須です")
	}

	// 名前検証
	if c.Name == "" {
		return NewErrInvalidInput("card.name", "カード名は必須です")
	}
	if len(c.Name) > MaxCardNameLength {
		return NewErrInvalidInput("card.name", fmt.Sprintf("カード名は%d文字以内である必要があります", MaxCardNameLength))
	}

	// タイプ検証
	if c.Type == "" {
		return NewErrInvalidInput("card.type", "カードタイプは必須です")
	}
	if c.Type != CardTypeUnit && c.Type != CardTypeSpell && c.Type != CardTypeLeader {
		return NewErrInvalidInput("card.type", "無効なカードタイプです")
	}

	// コスト検証
	if c.Cost < 0 {
		return NewErrInvalidInput("card.cost", "カードコストは0以上である必要があります")
	}
	if c.Cost > MaxCardCost {
		return NewErrInvalidInput("card.cost", fmt.Sprintf("カードコストは%d以下である必要があります", MaxCardCost))
	}

	// 効果テキスト検証
	if len(c.Effect) > MaxCardEffectLength {
		return NewErrInvalidInput("card.effect", fmt.Sprintf("効果テキストは%d文字以内である必要があります", MaxCardEffectLength))
	}

	// ユニットカードの場合、攻撃力と防御力が必要
	if c.Type == CardTypeUnit {
		if c.Attack == nil {
			return NewErrInvalidInput("card.attack", "ユニットカードには攻撃力が必要です")
		}
		if c.Defense == nil {
			return NewErrInvalidInput("card.defense", "ユニットカードには防御力が必要です")
		}
		if *c.Attack < 0 {
			return NewErrInvalidInput("card.attack", "攻撃力は0以上である必要があります")
		}
		if *c.Defense < 0 {
			return NewErrInvalidInput("card.defense", "防御力は0以上である必要があります")
		}
		if *c.Attack > MaxCardAttack {
			return NewErrInvalidInput("card.attack", fmt.Sprintf("攻撃力は%d以下である必要があります", MaxCardAttack))
		}
		if *c.Defense > MaxCardDefense {
			return NewErrInvalidInput("card.defense", fmt.Sprintf("防御力は%d以下である必要があります", MaxCardDefense))
		}

		// 特性の数を検証
		if len(c.Traits) > MaxCardTraits {
			return NewErrInvalidInput("card.traits", fmt.Sprintf("特性は%d個以内である必要があります", MaxCardTraits))
		}
	}

	// スペルまたはリーダーカードの場合、特性は設定できない
	if c.Type == CardTypeSpell || c.Type == CardTypeLeader {
		if len(c.Traits) > 0 {
			return NewErrInvalidInput("card.traits", "スペルおよびリーダーカードには特性を設定できません")
		}
	}

	// スペルまたはリーダーカードの場合、攻撃力と防御力は設定できない
	if c.Type == CardTypeSpell || c.Type == CardTypeLeader {
		if c.Attack != nil {
			return NewErrInvalidInput("card.attack", "スペルおよびリーダーカードには攻撃力を設定できません")
		}
		if c.Defense != nil {
			return NewErrInvalidInput("card.defense", "スペルおよびリーダーカードには防御力を設定できません")
		}
	}

	return nil
}

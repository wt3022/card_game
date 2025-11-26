package effect

import (
	"card_game/internal/core/entity"
)

// ========================================
// ユニット破壊ハンドラー
// OnDestroy効果を処理
// 設計方針:
// - Observerパターンの実装
// - 依存性逆転の原則（DIP）に従う
// ========================================

// DestructionHandler OnDestroy効果を処理するハンドラー
type DestructionHandler struct {
	processor *Processor
}

// NewDestructionHandler 破壊ハンドラーを作成
func NewDestructionHandler(processor *Processor) *DestructionHandler {
	return &DestructionHandler{
		processor: processor,
	}
}

// OnUnitDestroyed ユニット破壊時に呼ばれる
func (h *DestructionHandler) OnUnitDestroyed(unit *entity.Unit, owner *entity.Player) error {
	// ユニットに対応するカードを復元
	card := &entity.Card{
		ID:     unit.CardID,
		Name:   unit.Name,
		Type:   entity.CardTypeUnit,
		Cost:   unit.Cost,
		Effect: unit.Effect,
		Traits: unit.Traits,
	}

	if unit.Attack > 0 {
		attack := unit.Attack
		card.Attack = &attack
	}
	if unit.Defense > 0 {
		defense := unit.Defense
		card.Defense = &defense
	}

	// CardEffectが定義されている場合のみ処理
	if card.CardEffect != nil {
		// OnDestroy効果を処理
		return h.processor.ProcessTimingEffects(card, entity.EffectTimingOnDestroy, owner.ID, nil)
	}

	return nil
}

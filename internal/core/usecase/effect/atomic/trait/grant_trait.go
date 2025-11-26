package trait

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// 特性を付与
// ユニットに特性を付与
// ========================================

// ユニットに特性を付与
func ExecuteGrantTrait(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	// 付与する特性を取得
	var trait entity.Trait

	// サポートするパラメータ名を順にチェック（過去の定義により key が "keyword" または "trait" の場合がある）
	if v, exists := effect.Parameters["keyword"]; exists {
		switch tv := v.(type) {
		case string:
			trait = entity.Trait(tv)
		case entity.Trait:
			trait = tv
		default:
			return fmt.Errorf("invalid trait parameter")
		}
	} else if v, exists := effect.Parameters["trait"]; exists {
		switch tv := v.(type) {
		case string:
			trait = entity.Trait(tv)
		case entity.Trait:
			trait = tv
		default:
			return fmt.Errorf("invalid trait parameter")
		}
	} else {
		return fmt.Errorf("invalid trait parameter")
	}

	// 有効な特性か確認
	if !entity.IsValidTrait(trait) {
		return fmt.Errorf("invalid trait: %s", trait)
	}

	for _, target := range targets {
		unit, ok := target.(*entity.Unit)
		if !ok {
			continue
		}

		// 効果盾チェック：効果を受けない
		if unit.HasTrait(entity.TraitEffectShield) {
			game.AddLog(sourcePlayer.ID, "特性付与無効", fmt.Sprintf("%s は効果盾により特性付与を受けない", unit.Name))
			continue
		}

		// 既に持っている場合はスキップ
		if !unit.HasTrait(trait) {
			unit.AddTrait(trait)
			game.AddLog(sourcePlayer.ID, "特性付与",
				fmt.Sprintf("%s に %s を付与", unit.Name, entity.GetTraitName(trait)))
		}
	}

	return nil
}

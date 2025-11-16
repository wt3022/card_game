package atomic

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
	traitStr, ok := effect.Parameters["keyword"].(string)
	if !ok {
		return fmt.Errorf("invalid trait parameter")
	}

	trait := entity.Trait(traitStr)

	// 有効な特性か確認
	if !entity.IsValidTrait(trait) {
		return fmt.Errorf("invalid trait: %s", trait)
	}

	for _, target := range targets {
		unit, ok := target.(*entity.Unit)
		if !ok {
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

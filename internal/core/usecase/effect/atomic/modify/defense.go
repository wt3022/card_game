package modify

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// ユニットの守備力を増減
func ExecuteModifyDefense(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		unit, ok := target.(*entity.Unit)
		if !ok {
			continue
		}

		// 効果盾チェック：効果を受けない
		if unit.HasTrait(entity.TraitEffectShield) {
			game.AddLog(sourcePlayer.ID, "守備力変更無効", fmt.Sprintf("%s は効果盾により守備力変更を受けない", unit.Name))
			continue
		}

		unit.ModifyDefense(effect.Value)
		action := "増加"
		if effect.Value < 0 {
			action = "減少"
		}
		game.AddLog(sourcePlayer.ID, "守備力変更", fmt.Sprintf("%s の守備力を %d %s", unit.Name, absInt(effect.Value), action))
	}
	return nil
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

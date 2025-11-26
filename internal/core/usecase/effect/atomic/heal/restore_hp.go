package heal

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// ユニットまたはプレイヤーのHPを回復
func ExecuteRestoreHP(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		switch t := target.(type) {
		case *entity.Player:
			// プレイヤーのHP回復
			t.HealHP(effect.Value)
			game.AddLog(sourcePlayer.ID, "回復", fmt.Sprintf("%s のHPを %d 回復", t.Name, effect.Value))
		case *entity.Unit:
			// 効果盾チェック：効果を受けない
			if t.HasTrait(entity.TraitEffectShield) {
				game.AddLog(sourcePlayer.ID, "回復無効", fmt.Sprintf("%s は効果盾により回復を受けない", t.Name))
				continue
			}

			// ユニットの守備力回復
			t.Heal(effect.Value)
			game.AddLog(sourcePlayer.ID, "回復", fmt.Sprintf("%s の守備力を %d 回復", t.Name, effect.Value))
		}
	}
	return nil
}

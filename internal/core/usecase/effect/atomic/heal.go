package atomic

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// HP回復処理
// ユニットまたはプレイヤーのHPを回復
// ========================================

// ユニットまたはプレイヤーのHPを回復
func ExecuteRestoreHP(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		switch t := target.(type) {
		case *entity.Player:
			// プレイヤーのHP回復
			t.HealHP(effect.Value)
			game.AddLog(sourcePlayer.ID, "回復", fmt.Sprintf("%s のHPを %d 回復", t.Name, effect.Value))

		case *entity.Unit:
			// ユニットの守備力回復
			t.Heal(effect.Value)
			game.AddLog(sourcePlayer.ID, "回復", fmt.Sprintf("%s の守備力を %d 回復", t.Name, effect.Value))
		}
	}

	return nil
}

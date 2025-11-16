package atomic

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// ダメージ処理
// ユニットまたはプレイヤーにダメージを与える
// ========================================

// ユニットまたはプレイヤーにダメージを与える
func ExecuteDealDamage(effect *entity.AtomicEffect, sourcePlayer, opponent *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		switch t := target.(type) {
		case *entity.Player:
			// プレイヤーへのダメージ
			t.TakeDamage(effect.Value)
			game.AddLog(sourcePlayer.ID, "ダメージ", fmt.Sprintf("%s に %d ダメージ", t.Name, effect.Value))

		case *entity.Unit:
			// ユニットへの効果ダメージ
			destroyed := t.TakeDamage(effect.Value, true)
			game.AddLog(sourcePlayer.ID, "ダメージ", fmt.Sprintf("%s に %d ダメージ", t.Name, effect.Value))

			if destroyed {
				owner := game.GetPlayerByID(t.OwnerID)
				owner.RemoveUnitFromField(t.InstanceID)
				game.AddLog(t.OwnerID, "ユニット破壊", fmt.Sprintf("%s が破壊された", t.Name))
			}
		}
	}

	return nil
}

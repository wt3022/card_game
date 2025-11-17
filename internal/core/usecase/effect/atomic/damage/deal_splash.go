package damage

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// 範囲ダメージ（複数ターゲットに一括ダメージ）
func ExecuteDealSplash(effect *entity.AtomicEffect, sourcePlayer, opponent *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		switch t := target.(type) {
		case *entity.Player:
			t.TakeDamage(effect.Value)
			game.AddLog(sourcePlayer.ID, "範囲ダメージ", fmt.Sprintf("%s に %d ダメージ (範囲)", t.Name, effect.Value))
		case *entity.Unit:
			destroyed := t.TakeDamage(effect.Value, true)
			game.AddLog(sourcePlayer.ID, "範囲ダメージ", fmt.Sprintf("%s に %d ダメージ (範囲)", t.Name, effect.Value))
			if destroyed {
				owner := game.GetPlayerByID(t.OwnerID)
				owner.RemoveUnitFromField(t.InstanceID)
				game.AddLog(t.OwnerID, "ユニット破壊", fmt.Sprintf("%s が破壊された (範囲)", t.Name))
			}
		}
	}
	return nil
}

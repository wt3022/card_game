package heal

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// 完全回復
func ExecuteFullRestore(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, t := range targets {
		switch v := t.(type) {
		case *entity.Player:
			v.HP = v.MaxHP
			game.AddLog(v.ID, "完全回復", "HPを全回復")
		case *entity.Unit:
			v.CurrentDefense = v.Defense
			game.AddLog(v.OwnerID, "完全回復", v.Name+"の守備力を全回復")
		}
	}
	return nil
}

package modify

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// 最大HP変更
func ExecuteModifyMaxHP(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, t := range targets {
		unit, ok := t.(*entity.Unit)
		if !ok {
			continue
		}
		unit.Defense += effect.Value
		if unit.Defense < 0 {
			unit.Defense = 0
		}
		if unit.CurrentDefense > unit.Defense {
			unit.CurrentDefense = unit.Defense
		}
		game.AddLog(unit.OwnerID, "最大HP変更", unit.Name+"の最大HPを変更")
	}
	return nil
}

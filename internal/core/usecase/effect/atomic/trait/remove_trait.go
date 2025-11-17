package trait

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// 特性除去
func ExecuteRemoveTrait(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	traitStr, ok := effect.Parameters["trait"].(string)
	if !ok {
		return entity.NewErrEffectNotImplemented("REMOVE_TRAIT_param")
	}
	trait := entity.Trait(traitStr)
	for _, t := range targets {
		unit, ok := t.(*entity.Unit)
		if !ok {
			continue
		}
		if unit.HasTrait(trait) {
			unit.RemoveTrait(trait)
			game.AddLog(unit.OwnerID, "特性除去", unit.Name+"から"+traitStr+"を除去")
		}
	}
	return nil
}

package mana

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// マナ回復
func ExecuteRestoreMana(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	amount := effect.Value
	if amount <= 0 {
		amount = 1
	}
	sourcePlayer.AddMana(amount)
	game.AddLog(sourcePlayer.ID, "マナ回復", fmt.Sprintf("マナを%d回復", amount))
	return nil
}

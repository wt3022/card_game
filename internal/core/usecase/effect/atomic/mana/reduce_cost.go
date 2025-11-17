package mana

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// コスト減少
func ExecuteReduceCost(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	amount := effect.Value
	if amount == 0 {
		amount = 1
	}
	for i := range sourcePlayer.Hand {
		sourcePlayer.Hand[i].Cost -= amount
		if sourcePlayer.Hand[i].Cost < 0 {
			sourcePlayer.Hand[i].Cost = 0
		}
	}
	game.AddLog(sourcePlayer.ID, "コスト減少", "手札のコストを減少")
	return nil
}

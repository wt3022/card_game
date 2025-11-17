package draw

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// 手札を捨てる
func ExecuteDiscardCard(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	count := effect.Value
	if count <= 0 {
		count = 1
	}
	if len(sourcePlayer.Hand) < count {
		count = len(sourcePlayer.Hand)
	}
	for i := 0; i < count; i++ {
		card := sourcePlayer.Hand[0]
		sourcePlayer.Hand = sourcePlayer.Hand[1:]
		sourcePlayer.Graveyard = append(sourcePlayer.Graveyard, card)
		game.AddLog(sourcePlayer.ID, "手札破棄", fmt.Sprintf("%sを捨てた", card.Name))
	}
	return nil
}

package draw

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// カードドロー処理
func ExecuteDrawCard(effect *entity.AtomicEffect, sourcePlayer *entity.Player, game port.GameStateReader) error {
	drawCount := effect.Value
	if drawCount <= 0 {
		drawCount = 1
	}
	drawnCards, err := sourcePlayer.DrawCards(drawCount)
	if err != nil {
		return err
	}
	game.AddLog(sourcePlayer.ID, "ドロー", fmt.Sprintf("%d 枚ドロー", len(drawnCards)))
	return nil
}

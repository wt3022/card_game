package atomic

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// カードドロー処理
// 指定された枚数のカードを引く
// ========================================

// カードドロー処理
func ExecuteDrawCard(effect *entity.AtomicEffect, sourcePlayer *entity.Player, game port.GameStateReader) error {
	// ドロー枚数を取得
	drawCount := effect.Value
	if drawCount <= 0 {
		drawCount = 1
	}

	// カードをドロー
	drawnCards, err := sourcePlayer.DrawCards(drawCount)
	if err != nil {
		// デッキ切れなどのエラーを返す
		return err
	}

	game.AddLog(sourcePlayer.ID, "ドロー", fmt.Sprintf("%d 枚ドロー", len(drawnCards)))

	return nil
}

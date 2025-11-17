package mana

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// マナを増加
// プレイヤーのマナを増加
// ========================================

// プレイヤーのマナを増加
func ExecuteGainMana(effect *entity.AtomicEffect, sourcePlayer *entity.Player, game port.GameStateReader) error {
	// マナ量を取得
	manaAmount := effect.Value
	if manaAmount <= 0 {
		manaAmount = 1
	}

	// マナを追加
	sourcePlayer.AddMana(manaAmount)

	game.AddLog(sourcePlayer.ID, "マナ増加", fmt.Sprintf("マナを %d 増加", manaAmount))

	return nil
}

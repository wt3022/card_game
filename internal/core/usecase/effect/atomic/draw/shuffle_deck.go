package draw

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"math/rand"
)

// デッキシャッフル
func ExecuteShuffleDeck(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	n := len(sourcePlayer.Deck)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		sourcePlayer.Deck[i], sourcePlayer.Deck[j] = sourcePlayer.Deck[j], sourcePlayer.Deck[i]
	}
	game.AddLog(sourcePlayer.ID, "デッキシャッフル", "デッキをシャッフル")
	return nil
}

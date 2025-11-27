package game

import (
	"card_game/internal/core/entity"
	"math/rand"
)

// CoinToss コイントスを実行（50%の確率で表）
func CoinToss() bool {
	return rand.Intn(2) == 0
}

// デッキをシャッフルする
func ShuffleDeck(in []entity.Card) []entity.Card {
	if len(in) == 0 {
		return in
	}
	shuffled := make([]entity.Card, len(in))
	round := rand.Intn(len(in))
	for r := 0; r < round; r++ {
		perm := rand.Perm(len(in))
		for i, v := range perm {
			shuffled[i] = in[v]
		}
		copy(in, shuffled)
	}
	if round == 0 {
		perm := rand.Perm(len(in))
		for i, v := range perm {
			shuffled[i] = in[v]
		}
		copy(in, shuffled)
	}
	return in
}

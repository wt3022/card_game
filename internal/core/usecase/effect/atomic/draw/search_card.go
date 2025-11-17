package draw

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// デッキからサーチ
func ExecuteSearchCard(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	// サーチ対象の指定は effect.Parameters に任せる
	// サポートされるパラメータ: "card_id" (string), "card_name" (string), "card_type" (entity.CardType)
	count := effect.Value
	if count <= 0 {
		count = 1
	}

	found := 0
	// 検索関数
	match := func(c entity.Card) bool {
		if effect.Parameters == nil {
			return false
		}
		if v, ok := effect.Parameters["card_id"].(string); ok && v != "" {
			return c.ID == v
		}
		if v, ok := effect.Parameters["card_name"].(string); ok && v != "" {
			return c.Name == v
		}
		if v, ok := effect.Parameters["card_type"].(entity.CardType); ok {
			return c.Type == v
		}
		return false
	}

	// デッキを走査して該当カードを手札へ移動（最初に見つけたものから順に）
	movedIDs := map[string]struct{}{}
	for i := 0; i < len(sourcePlayer.Deck) && found < count; i++ {
		card := sourcePlayer.Deck[i]
		if match(card) {
			sourcePlayer.Hand = append(sourcePlayer.Hand, card)
			movedIDs[card.ID] = struct{}{}
			found++
			game.AddLog(sourcePlayer.ID, "サーチ", fmt.Sprintf("デッキから %s を手札に加えた", card.Name))
		}
	}

	// デッキを再構築（移動したカードを除去）
	if found > 0 {
		rebuilt := make([]entity.Card, 0, len(sourcePlayer.Deck)-found)
		for _, c := range sourcePlayer.Deck {
			if _, ok := movedIDs[c.ID]; ok {
				continue
			}
			rebuilt = append(rebuilt, c)
		}
		sourcePlayer.Deck = rebuilt
	}

	if found == 0 {
		game.AddLog(sourcePlayer.ID, "サーチ", "該当するカードが見つからなかった")
	}

	return nil
}

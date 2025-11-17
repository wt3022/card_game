package unit

import "card_game/internal/core/entity"

// findCardByID: プレイヤーのデッキ・墓地・手札からカードを検索
func findCardByID(player *entity.Player, cardID string) *entity.Card {
	for i := range player.Deck {
		if player.Deck[i].ID == cardID {
			return &player.Deck[i]
		}
	}
	for i := range player.Hand {
		if player.Hand[i].ID == cardID {
			return &player.Hand[i]
		}
	}
	for i := range player.Graveyard {
		if player.Graveyard[i].ID == cardID {
			return &player.Graveyard[i]
		}
	}
	return nil
}

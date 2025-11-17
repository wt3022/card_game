package unit

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ユニットを手札に戻す
func ExecuteReturnToHand(effect *entity.AtomicEffect, sourcePlayer *entity.Player, opponent *entity.Player, targets []any, game port.GameStateReader) error {
	for _, t := range targets {
		unit, ok := t.(*entity.Unit)
		if !ok {
			continue
		}
		owner := game.GetPlayerByID(unit.OwnerID)
		if owner == nil {
			continue
		}
		removed := owner.RemoveUnitFromField(unit.InstanceID)
		if removed != nil {
			card := findCardByID(owner, removed.CardID)
			if card == nil {
				card = &entity.Card{
					ID:      removed.CardID,
					Name:    removed.Name,
					Type:    entity.CardTypeUnit,
					Cost:    removed.Cost,
					Attack:  &removed.Attack,
					Defense: &removed.Defense,
					Effect:  removed.Effect,
					Traits:  removed.Traits,
				}
			}
			owner.Hand = append(owner.Hand, *card)
		}
	}
	return nil
}

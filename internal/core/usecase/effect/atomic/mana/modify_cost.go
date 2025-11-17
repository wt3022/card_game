package mana

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// コスト変更
func ExecuteModifyCost(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	amount := effect.Value
	if amount == 0 {
		return nil
	}

	// ターゲットが指定されていればそれらに対してコスト変更を試みる
	if len(targets) > 0 {
		for _, t := range targets {
			switch v := t.(type) {
			case *entity.Card:
				// カードが手札内にあれば変更
				for i := range sourcePlayer.Hand {
					if sourcePlayer.Hand[i].ID == v.ID {
						sourcePlayer.Hand[i].Cost += amount
						if sourcePlayer.Hand[i].Cost < 0 {
							sourcePlayer.Hand[i].Cost = 0
						}
						game.AddLog(sourcePlayer.ID, "コスト変更", fmt.Sprintf("%s のコストを %d した", sourcePlayer.Hand[i].Name, amount))
					}
				}
			case *entity.Unit:
				v.Cost += amount
				if v.Cost < 0 {
					v.Cost = 0
				}
				game.AddLog(sourcePlayer.ID, "コスト変更", fmt.Sprintf("%s のコストを %d した", v.Name, amount))
			}
		}
		return nil
	}

	// ターゲットがない場合、手札全体のコストを変更
	for i := range sourcePlayer.Hand {
		sourcePlayer.Hand[i].Cost += amount
		if sourcePlayer.Hand[i].Cost < 0 {
			sourcePlayer.Hand[i].Cost = 0
		}
	}
	if amount < 0 {
		game.AddLog(sourcePlayer.ID, "コスト変更", fmt.Sprintf("手札のコストを %d 減少", -amount))
	} else {
		game.AddLog(sourcePlayer.ID, "コスト変更", fmt.Sprintf("手札のコストを %d 増加", amount))
	}
	return nil
}

package heal

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// 完全回復
func ExecuteFullRestore(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, t := range targets {
		switch v := t.(type) {
		case *entity.Player:
			v.HP = v.MaxHP
			game.AddLog(v.ID, "完全回復", "HPを全回復")
		case *entity.Unit:
			// 効果盾チェック：効果を受けない
			if v.HasTrait(entity.TraitEffectShield) {
				game.AddLog(sourcePlayer.ID, "完全回復無効", v.Name+" は効果盾により回復を受けない")
				continue
			}

			v.CurrentDefense = v.Defense
			game.AddLog(v.OwnerID, "完全回復", v.Name+"の守備力を全回復")
		}
	}
	return nil
}

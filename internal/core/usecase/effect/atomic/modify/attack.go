package modify

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// ステータス変更処理
// ユニットの攻撃力・守備力を増減
// ========================================

// ユニットの攻撃力を増減
func ExecuteModifyAttack(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		unit, ok := target.(*entity.Unit)
		if !ok {
			continue
		}

		// 攻撃力を変更
		unit.ModifyAttack(effect.Value)

		action := "増加"
		if effect.Value < 0 {
			action = "減少"
		}

		game.AddLog(sourcePlayer.ID, "攻撃力変更",
			fmt.Sprintf("%s の攻撃力を %d %s", unit.Name, absInt(effect.Value), action))
	}

	return nil
}

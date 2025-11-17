package unit

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// ユニット破壊処理
// 指定されたユニットを破壊
// ========================================

// ユニット破壊処理
func ExecuteDestroyUnit(sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	for _, target := range targets {
		unit, ok := target.(*entity.Unit)
		if !ok {
			continue
		}

		// 破壊処理
		owner := game.GetPlayerByID(unit.OwnerID)
		if owner == nil {
			continue
		}

		removed := owner.RemoveUnitFromField(unit.InstanceID)
		if removed != nil {
			game.AddLog(sourcePlayer.ID, "ユニット破壊", fmt.Sprintf("%s を破壊", unit.Name))
		}
	}

	return nil
}

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

		// 対象保護チェック: 対象不可/効果盾の確認
		// 対象不可の場合は効果の対象にならない
		if unit.HasTrait(entity.TraitUntargetable) {
			game.AddLog(sourcePlayer.ID, "破壊失敗", fmt.Sprintf("%s は対象不可のため破壊できない", unit.Name))
			continue
		}

		// 効果盾 (EffectShield) を持つユニットは、効果による破壊を防ぐ
		if unit.HasTrait(entity.TraitEffectShield) {
			game.AddLog(sourcePlayer.ID, "破壊無効化", fmt.Sprintf("%s は効果盾により破壊を無効化", unit.Name))
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

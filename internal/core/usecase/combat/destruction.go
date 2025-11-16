package combat

import (
	"card_game/internal/core/entity"
)

// ========================================
// ユニット破壊処理
// 戦闘で破壊されたユニットを処理
// 設計方針:
// - フィールドから削除
// - 墓地に送る
// - 重複コードを削減
// ========================================

// ユニットを破壊してフィールドから削除し墓地に送る
func handleUnitDestruction(player *entity.Player, unit *entity.Unit) {
	if unit == nil {
		return
	}

	// フィールドから削除
	destroyed := player.RemoveUnitFromField(unit.InstanceID)
	if destroyed == nil {
		return
	}

	// 墓地に送る
	player.Graveyard = append(player.Graveyard, entity.Card{
		ID:   destroyed.CardID,
		Name: destroyed.Name,
		Type: entity.CardTypeUnit,
		Cost: destroyed.Cost,
	})
}

// 相討ちの場合の処理
func handleMutualDestruction(attacker, defender *entity.Player, attackerUnit, defenderUnit *entity.Unit) {
	handleUnitDestruction(defender, defenderUnit)
	handleUnitDestruction(attacker, attackerUnit)
}

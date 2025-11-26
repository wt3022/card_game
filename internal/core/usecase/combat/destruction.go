package combat

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// ユニット破壊処理
// 戦闘で破壊されたユニットを処理
// 設計方針:
// - フィールドから削除
// - 墓地に送る
// - OnDestroy効果を通知
// - 依存性逆転の原則（DIP）に従う
// ========================================

// UnitDestructor ユニット破壊処理の構造体
type UnitDestructor struct {
	notifier port.UnitDestructionNotifier
}

// NewUnitDestructor ユニット破壊処理を作成
func NewUnitDestructor(notifier port.UnitDestructionNotifier) *UnitDestructor {
	return &UnitDestructor{
		notifier: notifier,
	}
}

// ユニットを破壊してフィールドから削除し墓地に送る
func (d *UnitDestructor) handleUnitDestruction(player *entity.Player, unit *entity.Unit) error {
	if unit == nil {
		return nil
	}

	// フィールドから削除
	destroyed := player.RemoveUnitFromField(unit.InstanceID)
	if destroyed == nil {
		return nil
	}

	// 墓地に送る
	player.Graveyard = append(player.Graveyard, entity.Card{
		ID:   destroyed.CardID,
		Name: destroyed.Name,
		Type: entity.CardTypeUnit,
		Cost: destroyed.Cost,
	})

	// OnDestroy効果を通知
	if d.notifier != nil {
		if err := d.notifier.NotifyUnitDestruction(destroyed, player); err != nil {
			return err
		}
	}

	return nil
}

// 相討ちの場合の処理
func (d *UnitDestructor) handleMutualDestruction(attacker, defender *entity.Player, attackerUnit, defenderUnit *entity.Unit) error {
	if err := d.handleUnitDestruction(defender, defenderUnit); err != nil {
		return err
	}
	return d.handleUnitDestruction(attacker, attackerUnit)
}

// パッケージ外からアクセス可能な関数（後方互換性のため）
func handleUnitDestruction(player *entity.Player, unit *entity.Unit) {
	// 通知なしの簡易版（既存コードとの互換性用）
	destructor := &UnitDestructor{notifier: nil}
	_ = destructor.handleUnitDestruction(player, unit)
}

func handleMutualDestruction(attacker, defender *entity.Player, attackerUnit, defenderUnit *entity.Unit) {
	destructor := &UnitDestructor{notifier: nil}
	_ = destructor.handleMutualDestruction(attacker, defender, attackerUnit, defenderUnit)
}

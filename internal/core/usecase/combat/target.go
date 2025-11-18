package combat

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// 攻撃対象判定
// 攻撃可能な対象を判定
// 設計方針:
// - Guardian特性の処理
// - Pierce特性の処理
// - Direct特性の処理
// ========================================

// 攻撃可能なターゲットを取得
func GetAttackableTargets(g port.GameState, playerID string, attackerID string) ([]string, bool, error) {
	player := g.GetPlayerByID(playerID)
	if player == nil {
		return nil, false, fmt.Errorf("プレイヤーが見つかりません")
	}

	attackerUnit := player.GetUnitByInstanceID(attackerID)
	if attackerUnit == nil {
		return nil, false, fmt.Errorf("ユニットが見つかりません")
	}

	opponent := g.GetOpponent(playerID)
	targets := []string{}
	canDirectAttack := false

	// Guardianユニットの確認
	guardians := g.GetGuardianUnits(opponent)
	if len(guardians) > 0 {
		// Guardianがいる場合の処理
		if canBypassGuardian(attackerUnit) {
			// Guardian無視可能：全ユニットが対象
			for i := range opponent.Field {
				targets = append(targets, opponent.Field[i].InstanceID)
			}
			canDirectAttack = true
		} else {
			// Guardianのみが対象
			targets = guardians
			canDirectAttack = false
		}
	} else {
		// Guardianがいない場合の処理
		if len(opponent.Field) == 0 {
			// フィールドが空でも、特性によっては直接攻撃不可（例: Charge はユニットにのみ攻撃可能）
			canDirectAttack = canAttackDirectly(opponent, attackerUnit)
		} else {
			// ユニットが対象
			for i := range opponent.Field {
				targets = append(targets, opponent.Field[i].InstanceID)
			}
			// 直接攻撃可能かチェック
			canDirectAttack = canAttackDirectly(opponent, attackerUnit)
		}
	}

	return targets, canDirectAttack, nil
}

// ========================================
// 判定関数
// ========================================

// プレイヤーへの直接攻撃が可能か判定
func canAttackDirectly(opponent *entity.Player, attackerUnit *entity.Unit) bool {
	// フィールドが空なら基本的に可能。ただし突進(Charge)はユニットにのみ攻撃可能なので例外とする
	if len(opponent.Field) == 0 {
		if attackerUnit.HasTrait(entity.TraitCharge) {
			return false
		}
		return true
	}

	// Rush特性を持つ場合は常にリーダーに攻撃可能
	if attackerUnit.HasTrait(entity.TraitRush) {
		return true
	}

	// Pierce特性を持つ場合も可能（Guardian無視として機能）
	if attackerUnit.HasTrait(entity.TraitPierce) {
		return true
	}

	return false
}

// Guardianを無視できるか判定
func canBypassGuardian(attackerUnit *entity.Unit) bool {
	// Pierce特性を持つ場合はGuardian無視
	return attackerUnit.HasTrait(entity.TraitPierce)
}

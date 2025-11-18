package combat

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// 戦闘処理
// ユニットの攻撃とダメージ計算を実行
// 設計方針:
// - 直接攻撃とユニット戦闘を分離
// - 特性処理を適用
// - 破壊処理を統一
// ========================================

// 攻撃を実行
func ExecuteAttack(g port.GameState, action entity.AttackAction) (*entity.CombatResult, error) {
	// 攻撃者のプレイヤーを取得
	attacker := g.GetPlayerByID(action.PlayerID)
	if attacker == nil {
		return nil, entity.NewErrNotFound("player", action.PlayerID)
	}

	// 攻撃者のユニットを取得
	attackerUnit := attacker.GetUnitByInstanceID(action.AttackerID)
	if attackerUnit == nil {
		return nil, entity.NewErrNotFound("unit", action.AttackerID)
	}

	// 攻撃可能かチェック
	if !attackerUnit.CanAttack() {
		return nil, entity.NewErrInvalidAction("attack", fmt.Sprintf("unit %s cannot attack", attackerUnit.Name))
	}

	opponent := g.GetOpponent(action.PlayerID)

	// 直接攻撃の場合
	if action.TargetID == nil {
		return executeDirectAttack(g, attacker, opponent, attackerUnit, action)
	}

	// ユニット対ユニットの戦闘
	defenderUnit := opponent.GetUnitByInstanceID(*action.TargetID)
	if defenderUnit == nil {
		return nil, entity.NewErrNotFound("unit", *action.TargetID)
	}

	return executeUnitCombat(g, attacker, opponent, attackerUnit, defenderUnit, action)
}

// ========================================
// 直接攻撃
// ========================================

// プレイヤーへの直接攻撃を実行
func executeDirectAttack(g port.GameState, _, opponent *entity.Player, attackerUnit *entity.Unit, action entity.AttackAction) (*entity.CombatResult, error) {
	// 直接攻撃可能かチェック
	if !canAttackDirectly(opponent, attackerUnit) {
		return nil, entity.NewErrInvalidAction("direct_attack", "直接攻撃できません")
	}

	// ダメージを与える
	damage := attackerUnit.GetAttack()
	opponent.TakeDamage(damage)

	// 攻撃使用済みにする
	attackerUnit.UseAttack()

	// ログ出力
	logger := g.GetLogger()
	logger.Info("「%s」が%sに直接攻撃！ (ダメージ: %d)", attackerUnit.Name, opponent.Name, damage)
	logger.Info("→ %s の HP: %d/%d", opponent.Name, opponent.HP, opponent.MaxHP)

	g.AddLog(action.PlayerID, "戦闘",
		fmt.Sprintf("%s が %s に直接攻撃 (%dダメージ)", attackerUnit.Name, opponent.Name, damage))

	// 勝利判定
	g.CheckVictoryConditions()

	return &entity.CombatResult{
		DirectDamage: damage,
	}, nil
}

// ========================================
// ユニット戦闘
// ========================================

// ユニット対ユニットの戦闘を実行
func executeUnitCombat(g port.GameState, attacker, defender *entity.Player, attackerUnit, defenderUnit *entity.Unit, action entity.AttackAction) (*entity.CombatResult, error) {
	// 守護ユニットの確認
	if g.HasGuardianUnits(defender) && !defenderUnit.HasTrait(entity.TraitGuardian) {
		return nil, entity.NewErrInvalidAction("attack", "守護ユニットが存在するため、先に守護ユニットを攻撃する必要があります")
	}

	logger := g.GetLogger()
	logger.Info("戦闘: 「%s」(攻:%d/守:%d) vs 「%s」(攻:%d/守:%d)",
		attackerUnit.Name, attackerUnit.GetAttack(), attackerUnit.GetCurrentDefense(),
		defenderUnit.Name, defenderUnit.GetAttack(), defenderUnit.GetCurrentDefense())

	// ダメージ計算
	attackerDamage := defenderUnit.GetAttack()
	defenderDamage := attackerUnit.GetAttack()

	// 貫通(Pierce)のために、防御側の事前守備値を保持
	defenderPreDefense := defenderUnit.GetCurrentDefense()

	// 両方同時にダメージを受ける（戦闘ダメージ）
	defenderDestroyed := defenderUnit.TakeDamage(defenderDamage, false)
	attackerDestroyed := attackerUnit.TakeDamage(attackerDamage, false)

	// 貫通処理: 攻撃者がPierceを持っていて、防御側が破壊された場合、余剰ダメージをプレイヤーへ与える
	if defenderDestroyed && attackerUnit.HasTrait(entity.TraitPierce) {
		overflow := defenderDamage - defenderPreDefense
		if overflow > 0 {
			opponent := g.GetOpponent(action.PlayerID)
			opponent.TakeDamage(overflow)
			logger.Info("貫通: %s の余剰ダメージ %d が %s に入った", attackerUnit.Name, overflow, opponent.Name)
			g.AddLog(action.PlayerID, "貫通", fmt.Sprintf("%s の余剰ダメージ %d が %s に入る", attackerUnit.Name, overflow, opponent.Name))
			// 勝利判定
			g.CheckVictoryConditions()
		}
	}

	// 結果を処理
	handleCombatResult(g, attacker, defender, attackerUnit, defenderUnit, attackerDestroyed, defenderDestroyed, attackerDamage, defenderDamage, action)

	// 攻撃使用済みにする（破壊されていなければ）
	if !attackerDestroyed {
		attackerUnit.UseAttack()
	}

	return &entity.CombatResult{
		Damage:            defenderDamage,
		DefenderDestroyed: defenderDestroyed,
		AttackerDestroyed: attackerDestroyed,
	}, nil
}

// 戦闘結果を処理
func handleCombatResult(g port.GameState, attacker, defender *entity.Player, attackerUnit, defenderUnit *entity.Unit, attackerDestroyed, defenderDestroyed bool, attackerDamage, defenderDamage int, action entity.AttackAction) {
	logger := g.GetLogger()

	if attackerDestroyed && defenderDestroyed {
		// 相討ち
		logger.Info("→ 相討ち！両方のユニットが破壊されました")
		handleMutualDestruction(attacker, defender, attackerUnit, defenderUnit)
		g.AddLog(action.PlayerID, "戦闘",
			fmt.Sprintf("%s と %s が相討ち", attackerUnit.Name, defenderUnit.Name))

	} else if defenderDestroyed {
		// 防御側が破壊
		logger.Info("→ 「%s」を破壊！", defenderUnit.Name)
		logger.Info("→ 「%s」の守備力: %d → %d",
			attackerUnit.Name, attackerUnit.GetCurrentDefense()+attackerDamage, attackerUnit.GetCurrentDefense())
		handleUnitDestruction(defender, defenderUnit)
		g.AddLog(action.PlayerID, "戦闘",
			fmt.Sprintf("%s が %s を破壊（反撃で%dダメージ受ける）", attackerUnit.Name, defenderUnit.Name, attackerDamage))

	} else if attackerDestroyed {
		// 攻撃側が破壊
		logger.Info("→ 「%s」を破壊！", attackerUnit.Name)
		logger.Info("→ 「%s」の守備力: %d → %d",
			defenderUnit.Name, defenderUnit.GetCurrentDefense()+defenderDamage, defenderUnit.GetCurrentDefense())
		handleUnitDestruction(attacker, attackerUnit)
		g.AddLog(action.PlayerID, "戦闘",
			fmt.Sprintf("%s の攻撃で %s が反撃により破壊される", defenderUnit.Name, attackerUnit.Name))

	} else {
		// どちらも破壊されない
		logger.Info("→ 「%s」の守備力: %d → %d",
			defenderUnit.Name, defenderUnit.GetCurrentDefense()+defenderDamage, defenderUnit.GetCurrentDefense())
		logger.Info("→ 「%s」の守備力: %d → %d",
			attackerUnit.Name, attackerUnit.GetCurrentDefense()+attackerDamage, attackerUnit.GetCurrentDefense())
		g.AddLog(action.PlayerID, "戦闘",
			fmt.Sprintf("%s が %s に攻撃（相互に%d/%dダメージ）", attackerUnit.Name, defenderUnit.Name, defenderDamage, attackerDamage))
	}
}

package effect

import (
	"card_game/internal/core/entity"
)

// ========================================
// 条件評価
// ゲーム状態に基づいて条件を評価
// ========================================

// 条件を評価
func (p *Processor) evaluateCondition(cond *entity.Condition, ctx *ExecutionContext) bool {
	if cond == nil {
		return true
	}

	sourcePlayer := p.game.GetPlayerByID(ctx.SourcePlayerID)
	if sourcePlayer == nil {
		return false
	}

	// 実際の値を取得
	actualValue := p.getConditionValue(cond.Type, sourcePlayer, ctx)

	// 比較演算を実行
	return p.compareValues(actualValue, cond.Operator, cond.Value)
}

// 条件タイプに基づいて実際の値を取得
func (p *Processor) getConditionValue(condType entity.ConditionType, sourcePlayer *entity.Player, ctx *ExecutionContext) int {
	switch condType {
	case entity.ConditionPlayerHP:
		return sourcePlayer.HP

	case entity.ConditionPlayerMana:
		return sourcePlayer.CurrentTurnMana

	case entity.ConditionUnitCount:
		return len(sourcePlayer.Field)

	case entity.ConditionHandSize:
		return len(sourcePlayer.Hand)

	case entity.ConditionDeckSize:
		return len(sourcePlayer.Deck)

	case entity.ConditionTurnNumber:
		// ターン数は変数から取得
		if turn, ok := ctx.Variables["turn_number"]; ok {
			return turn
		}
		return 0

	case entity.ConditionUnitAttack:
		// 特定ユニットの攻撃力（TargetIDが必要）
		if ctx.TargetID != nil {
			if unit := sourcePlayer.GetUnitByInstanceID(*ctx.TargetID); unit != nil {
				return unit.GetAttack()
			}
		}
		return 0

	case entity.ConditionUnitDefense:
		// 特定ユニットの防御力（TargetIDが必要）
		if ctx.TargetID != nil {
			if unit := sourcePlayer.GetUnitByInstanceID(*ctx.TargetID); unit != nil {
				return unit.GetDefense()
			}
		}
		return 0

	case entity.ConditionCardPlayed:
		// このターン使用したカード数（変数から取得）
		if count, ok := ctx.Variables["cards_played_this_turn"]; ok {
			return count
		}
		return 0

	case entity.ConditionDamageTaken:
		// 受けたダメージ量（変数から取得）
		if damage, ok := ctx.Variables["damage_taken"]; ok {
			return damage
		}
		return 0

	default:
		return 0
	}
}

// 比較演算を実行
func (p *Processor) compareValues(actual int, operator entity.ComparisonOperator, expected int) bool {
	switch operator {
	case entity.OperatorEqual:
		return actual == expected

	case entity.OperatorNotEqual:
		return actual != expected

	case entity.OperatorLessThan:
		return actual < expected

	case entity.OperatorGreaterThan:
		return actual > expected

	case entity.OperatorLessThanOrEqual:
		return actual <= expected

	case entity.OperatorGreaterThanOrEqual:
		return actual >= expected

	default:
		return false
	}
}

// ========================================
// ヘルパー関数
// ========================================

// 条件が満たされているか確認（公開用）
func (p *Processor) IsConditionMet(cond *entity.Condition, sourcePlayerID string, targetID *string) bool {
	ctx := &ExecutionContext{
		SourcePlayerID: sourcePlayerID,
		TargetID:       targetID,
		Variables:      make(map[string]int),
	}
	return p.evaluateCondition(cond, ctx)
}

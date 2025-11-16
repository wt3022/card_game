package effect

import (
	"math/rand"

	"card_game/internal/core/entity"
)

// ========================================
// 対象解決とフィルタリング
// 効果の対象を動的に解決し、条件に基づいてフィルタリング
// 設計方針:
// - プレイヤー、ユニット、ランダムなど多様な対象指定
// - フィルタによる絞り込み（攻撃力、守備力、特性など）
// - ランダム選択と個数制限
// ========================================

// 対象セレクターから実際の対象を解決
func (p *Processor) resolveTargets(selector entity.TargetSelector, ctx *ExecutionContext) []interface{} {
	sourcePlayer := p.game.GetPlayerByID(ctx.SourcePlayerID)
	opponent := p.game.GetOpponent(ctx.SourcePlayerID)

	// 基本対象を解決
	targets := p.resolveBaseTargets(selector.Type, sourcePlayer, opponent, ctx)

	// フィルタを適用
	if selector.Filter != nil {
		targets = p.applyFilter(targets, selector.Filter)
	}

	// 個数制限を適用
	targets = p.applyCountLimit(targets, selector.Count, selector.Random)

	return targets
}

// ========================================
// 基本対象解決
// ========================================

// 基本的な対象タイプから対象を解決
func (p *Processor) resolveBaseTargets(targetType entity.EffectTarget, sourcePlayer, opponent *entity.Player, ctx *ExecutionContext) []interface{} {
	targets := []interface{}{}

	switch targetType {
	case entity.EffectTargetSelf:
		// 自分自身（プレイヤー）
		targets = append(targets, sourcePlayer)

	case entity.EffectTargetOpponent:
		// 相手プレイヤー
		targets = append(targets, opponent)

	case entity.EffectTargetAllies:
		// 味方ユニット全て
		for i := range sourcePlayer.Field {
			targets = append(targets, &sourcePlayer.Field[i])
		}

	case entity.EffectTargetEnemies:
		// 敵ユニット全て
		for i := range opponent.Field {
			targets = append(targets, &opponent.Field[i])
		}

	case entity.EffectTargetAllUnits:
		// すべてのユニット
		for i := range sourcePlayer.Field {
			targets = append(targets, &sourcePlayer.Field[i])
		}
		for i := range opponent.Field {
			targets = append(targets, &opponent.Field[i])
		}

	case entity.EffectTargetRandomAlly:
		// ランダムな味方ユニット1体
		if len(sourcePlayer.Field) > 0 {
			idx := rand.Intn(len(sourcePlayer.Field))
			targets = append(targets, &sourcePlayer.Field[idx])
		}

	case entity.EffectTargetRandomEnemy:
		// ランダムな敵ユニット1体
		if len(opponent.Field) > 0 {
			idx := rand.Intn(len(opponent.Field))
			targets = append(targets, &opponent.Field[idx])
		}

	case entity.EffectTargetSpecific:
		// 特定の対象（TargetIDで指定）
		if ctx.TargetID != nil {
			if unit := sourcePlayer.GetUnitByInstanceID(*ctx.TargetID); unit != nil {
				targets = append(targets, unit)
			} else if unit := opponent.GetUnitByInstanceID(*ctx.TargetID); unit != nil {
				targets = append(targets, unit)
			}
		}
	}

	return targets
}

// ========================================
// フィルタリング
// ========================================

// フィルタを適用してユニットを絞り込む
func (p *Processor) applyFilter(targets []interface{}, filter *entity.TargetFilter) []interface{} {
	if filter == nil {
		return targets
	}

	filtered := []interface{}{}

	for _, target := range targets {
		unit, ok := target.(*entity.Unit)
		if !ok {
			// ユニット以外はフィルタリングしない
			continue
		}

		// すべてのフィルタ条件を満たすかチェック
		if p.matchesFilter(unit, filter) {
			filtered = append(filtered, unit)
		}
	}

	return filtered
}

// ユニットがフィルタ条件を満たすか判定
func (p *Processor) matchesFilter(unit *entity.Unit, filter *entity.TargetFilter) bool {
	// 攻撃力フィルタ
	if filter.MinAttack != nil && unit.Attack < *filter.MinAttack {
		return false
	}
	if filter.MaxAttack != nil && unit.Attack > *filter.MaxAttack {
		return false
	}

	// 防御力フィルタ
	if filter.MinDefense != nil && unit.Defense < *filter.MinDefense {
		return false
	}
	if filter.MaxDefense != nil && unit.Defense > *filter.MaxDefense {
		return false
	}

	// コストフィルタ
	if filter.MinCost != nil && unit.Cost < *filter.MinCost {
		return false
	}
	if filter.MaxCost != nil && unit.Cost > *filter.MaxCost {
		return false
	}

	// 特性フィルタ（すべて持っている必要がある）
	if len(filter.HasTrait) > 0 {
		if !unit.HasAllTraits(filter.HasTrait) {
			return false
		}
	}

	// 特性欠如フィルタ（いずれかを持っていない必要がある）
	if len(filter.LackTrait) > 0 {
		if unit.HasAnyTrait(filter.LackTrait) {
			return false
		}
	}

	return true
}

// ========================================
// 個数制限
// ========================================

// 個数制限とランダム選択を適用
func (p *Processor) applyCountLimit(targets []interface{}, count int, random bool) []interface{} {
	if count <= 0 || len(targets) == 0 {
		return targets
	}

	// 対象数が上限以下の場合はそのまま返す
	if len(targets) <= count {
		return targets
	}

	// ランダム選択
	if random {
		rand.Shuffle(len(targets), func(i, j int) {
			targets[i], targets[j] = targets[j], targets[i]
		})
	}

	// 個数制限を適用
	return targets[:count]
}

package effect

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/core/usecase/effect/atomic"
)

// ========================================
// 効果処理エンジン
// コンポーザブル効果システムの実装
// 設計方針:
// - 原子効果と演算子の組み合わせで複雑な効果を表現
// - 効果定義はツリー構造で管理
// - 条件評価と対象解決を分離
// - 原子効果の実装はatomicパッケージに委譲
// ========================================

// コンポーザブル効果処理エンジン
type Processor struct {
	game port.GameStateReader
}

// 新しい効果処理エンジンを作成
func NewProcessor(g port.GameStateReader) *Processor {
	return &Processor{game: g}
}

// 効果実行時のコンテキスト
type ExecutionContext struct {
	SourcePlayerID string         // 効果の発動元プレイヤーID
	TargetID       *string        // 対象ID（指定された場合）
	Variables      map[string]int // 実行中の変数（破壊数、ダメージ量など）
}

// ========================================
// 公開API
// ========================================

// 効果定義を処理
func (p *Processor) ProcessEffectDefinition(def *entity.EffectDefinition, sourcePlayerID string, targetID *string) error {
	if def == nil || def.Root == nil {
		return nil
	}

	ctx := &ExecutionContext{
		SourcePlayerID: sourcePlayerID,
		TargetID:       targetID,
		Variables:      make(map[string]int),
	}

	return p.executeNode(def.Root, ctx)
}

// 特定のタイミングで発動する効果を処理
func (p *Processor) ProcessTimingEffects(card *entity.Card, timing entity.EffectTiming, playerID string, targetID *string) error {
	if card == nil || card.CardEffect == nil {
		return nil
	}

	for _, def := range card.CardEffect.Definitions {
		if def.Root != nil {
			// ルートノードから最初の効果のTimingを取得
			if effect := p.getFirstEffect(def.Root); effect != nil && effect.Timing == timing {
				if err := p.ProcessEffectDefinition(def, playerID, targetID); err != nil {
					return fmt.Errorf("効果処理エラー (%s): %w", timing, err)
				}
			}
		}
	}
	return nil
}

// ========================================
// 原子効果実行
// ========================================

// 原子効果を実行
func (p *Processor) executeAtomicEffect(effect *entity.AtomicEffect, ctx *ExecutionContext) error {
	// 条件チェック
	if effect.Condition != nil && !p.evaluateCondition(effect.Condition, ctx) {
		return nil
	}

	sourcePlayer := p.game.GetPlayerByID(ctx.SourcePlayerID)
	opponent := p.game.GetOpponent(ctx.SourcePlayerID)

	// 対象を解決
	targets := p.resolveTargets(effect.Target, ctx)

	// 原子効果タイプに応じて処理を委譲
	switch effect.Type {
	case entity.AtomicEffectDealDamage:
		return atomic.ExecuteDealDamage(effect, sourcePlayer, opponent, targets, p.game)

	case entity.AtomicEffectRestoreHP:
		return atomic.ExecuteRestoreHP(effect, sourcePlayer, targets, p.game)

	case entity.AtomicEffectDrawCard:
		return atomic.ExecuteDrawCard(effect, sourcePlayer, p.game)

	case entity.AtomicEffectModifyAttack:
		return atomic.ExecuteModifyAttack(effect, sourcePlayer, targets, p.game)

	case entity.AtomicEffectModifyDefense:
		return atomic.ExecuteModifyDefense(effect, sourcePlayer, targets, p.game)

	case entity.AtomicEffectDestroyUnit:
		return atomic.ExecuteDestroyUnit(sourcePlayer, targets, p.game)

	case entity.AtomicEffectGrantTrait:
		return atomic.ExecuteGrantTrait(effect, sourcePlayer, targets, p.game)

	case entity.AtomicEffectGainMana:
		return atomic.ExecuteGainMana(effect, sourcePlayer, p.game)

	case entity.AtomicEffectSummonUnit:
		// NOTE: 効果によるユニット召喚（トークン生成など）は未実装
		return entity.NewErrEffectNotImplemented(string(effect.Type))

	default:
		return entity.NewErrEffectNotImplemented(string(effect.Type))
	}
}

// ========================================
// ヘルパー関数
// ========================================

// ノードから最初の効果を取得
func (p *Processor) getFirstEffect(node *entity.EffectChainNode) *entity.AtomicEffect {
	if node == nil {
		return nil
	}

	switch node.Type {
	case entity.OperatorSequential:
		if seq, ok := node.GetSequential(); ok {
			if seq.Effect != nil {
				return seq.Effect
			}
			return p.getFirstEffect(seq.Next)
		}

	case entity.OperatorParallel:
		if par, ok := node.GetParallel(); ok && len(par.Children) > 0 {
			return p.getFirstEffect(par.Children[0])
		}

	case entity.OperatorIfElse:
		if ifElse, ok := node.GetIfElse(); ok {
			return p.getFirstEffect(ifElse.Then)
		}

	case entity.OperatorRepeat:
		if repeat, ok := node.GetRepeat(); ok {
			return p.getFirstEffect(repeat.Effect)
		}

	case entity.OperatorForEach:
		if forEach, ok := node.GetForEach(); ok {
			return p.getFirstEffect(forEach.Effect)
		}
	}

	return nil
}

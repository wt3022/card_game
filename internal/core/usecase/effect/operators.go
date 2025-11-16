package effect

import (
	"fmt"

	"card_game/internal/core/entity"
)

// ========================================
// 演算子実行
// 効果チェーンの演算子を実行
// 設計方針:
// - 各演算子の実行ロジックを分離
// - Sequential, Parallel, IfElse, Repeat, ForEach, Choiceに対応
// ========================================

// ノードを実行（演算子別の処理）
func (p *Processor) executeNode(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	if node == nil {
		return nil
	}

	switch node.Type {
	case entity.OperatorSequential:
		return p.executeSequential(node, ctx)

	case entity.OperatorParallel:
		return p.executeParallel(node, ctx)

	case entity.OperatorIfElse:
		return p.executeIfElse(node, ctx)

	case entity.OperatorRepeat:
		return p.executeRepeat(node, ctx)

	case entity.OperatorForEach:
		return p.executeForEach(node, ctx)

	case entity.OperatorChoice:
		return p.executeChoice(node, ctx)

	default:
		return fmt.Errorf("unknown operator type: %s", node.Type)
	}
}

// ========================================
// 各演算子の実行
// ========================================

// 順次実行: Aが完了したらBを実行
func (p *Processor) executeSequential(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	seq, ok := node.GetSequential()
	if !ok || seq == nil {
		return fmt.Errorf("invalid sequential node: node is nil or type mismatch")
	}

	// 現在の効果を実行（省略可能）
	if seq.Effect != nil {
		if err := p.executeAtomicEffect(seq.Effect, ctx); err != nil {
			return err
		}
	}

	// 次のノードを実行（nilの場合は終了）
	if seq.Next != nil {
		return p.executeNode(seq.Next, ctx)
	}

	return nil
}

// 並列実行: AとBを独立して実行
func (p *Processor) executeParallel(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	par, ok := node.GetParallel()
	if !ok || par == nil {
		return fmt.Errorf("invalid parallel node: node is nil or type mismatch")
	}

	// 全ての子ノードを実行
	for _, child := range par.Children {
		if err := p.executeNode(child, ctx); err != nil {
			return err
		}
	}

	// 並列実行後の次のノードを実行（省略可能）
	if par.Next != nil {
		return p.executeNode(par.Next, ctx)
	}

	return nil
}

// 条件分岐: IF 条件 THEN A ELSE B
func (p *Processor) executeIfElse(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	ifElse, ok := node.GetIfElse()
	if !ok || ifElse == nil {
		return fmt.Errorf("invalid if_else node: node is nil or type mismatch")
	}

	if ifElse.Condition == nil {
		return fmt.Errorf("invalid if_else node: condition is nil")
	}

	// 条件を評価
	if p.evaluateCondition(ifElse.Condition, ctx) {
		// 条件が真: Thenを実行
		return p.executeNode(ifElse.Then, ctx)
	} else if ifElse.Else != nil {
		// 条件が偽: Elseを実行（省略可能）
		return p.executeNode(ifElse.Else, ctx)
	}

	return nil
}

// 繰り返し: REPEAT(A, N回)
func (p *Processor) executeRepeat(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	repeat, ok := node.GetRepeat()
	if !ok || repeat == nil {
		return fmt.Errorf("invalid repeat node: node is nil or type mismatch")
	}

	if repeat.Count <= 0 {
		return fmt.Errorf("invalid repeat node: count must be positive")
	}

	// 指定回数繰り返し実行
	for i := 0; i < repeat.Count; i++ {
		if err := p.executeNode(repeat.Effect, ctx); err != nil {
			return err
		}
	}

	return nil
}

// 反復: FOREACH 対象集合 DO A
func (p *Processor) executeForEach(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	forEach, ok := node.GetForEach()
	if !ok || forEach == nil {
		return fmt.Errorf("invalid for_each node: node is nil or type mismatch")
	}

	// 対象を解決
	targets := p.resolveTargets(forEach.Target, ctx)

	// 各対象に対して実行
	for range targets {
		// 各対象用の新しいコンテキスト
		newCtx := &ExecutionContext{
			SourcePlayerID: ctx.SourcePlayerID,
			TargetID:       ctx.TargetID,
			Variables:      ctx.Variables,
		}
		if err := p.executeNode(forEach.Effect, newCtx); err != nil {
			return err
		}
	}

	return nil
}

// 選択: プレイヤーが選択肢から選択（未実装）
func (p *Processor) executeChoice(node *entity.EffectChainNode, ctx *ExecutionContext) error {
	return fmt.Errorf("choice operator is not implemented yet")
}

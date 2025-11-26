package converter

import (
	cardgamev1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/internal/core/entity"
	"fmt"
)

// ========================================
// Proto → Entity 変換
// Protocol Buffersメッセージをドメインエンティティに変換
// 設計方針:
// - バリデーション: 不正な値は適切にエラー処理
// - Nil安全: nilチェックを徹底
// - デフォルト値: UNSPECIFIED値を適切に処理
// ========================================

// CardFromProto ProtoからCardに変換
func CardFromProto(pb *cardgamev1.Card) *entity.Card {
	if pb == nil {
		return nil
	}

	return &entity.Card{
		ID:      pb.Id,
		Name:    pb.Name,
		Type:    CardTypeFromProto(pb.Type),
		Cost:    int(pb.Cost),
		Attack:  optionalInt32ToPtrInt(pb.Attack),
		Defense: optionalInt32ToPtrInt(pb.Defense),
		Effect:  pb.Effect,
		Traits:  TraitsFromProto(pb.Traits),
		// CardEffectは設定しない（サーバー側でのみ管理）
	}
}

// UnitFromProto ProtoからUnitに変換
func UnitFromProto(pb *cardgamev1.Unit) *entity.Unit {
	if pb == nil {
		return nil
	}

	return &entity.Unit{
		CardID:           pb.CardId,
		InstanceID:       pb.InstanceId,
		Name:             pb.Name,
		Cost:             int(pb.Cost),
		Attack:           int(pb.Attack),
		Defense:          int(pb.Defense),
		CurrentDefense:   int(pb.CurrentDefense),
		Traits:           TraitsFromProto(pb.Traits),
		Effect:           pb.Effect,
		AttacksRemaining: int(pb.AttacksRemaining),
		SummonedThisTurn: pb.SummonedThisTurn,
		OwnerID:          pb.OwnerId,
	}
}

// ========================================
// Enum変換
// ========================================

// CardTypeFromProto CardType変換
func CardTypeFromProto(pbType cardgamev1.CardType) entity.CardType {
	switch pbType {
	case cardgamev1.CardType_CARD_TYPE_UNIT:
		return entity.CardTypeUnit
	case cardgamev1.CardType_CARD_TYPE_SPELL:
		return entity.CardTypeSpell
	case cardgamev1.CardType_CARD_TYPE_LEADER:
		return entity.CardTypeLeader
	default:
		return entity.CardTypeUnit // デフォルト
	}
}

// TraitFromProto Trait変換
func TraitFromProto(pbTrait cardgamev1.Trait) entity.Trait {
	switch pbTrait {
	case cardgamev1.Trait_TRAIT_RUSH:
		return entity.TraitRush
	case cardgamev1.Trait_TRAIT_CHARGE:
		return entity.TraitCharge
	case cardgamev1.Trait_TRAIT_WINDFURY:
		return entity.TraitWindfury
	case cardgamev1.Trait_TRAIT_PIERCE:
		return entity.TraitPierce
	case cardgamev1.Trait_TRAIT_GUARDIAN:
		return entity.TraitGuardian
	case cardgamev1.Trait_TRAIT_EFFECT_SHIELD:
		return entity.TraitEffectShield
	case cardgamev1.Trait_TRAIT_UNTARGETABLE:
		return entity.TraitUntargetable
	default:
		return "" // 不明な特性は空文字列
	}
}

// TraitsFromProto Traitリスト変換
func TraitsFromProto(pbTraits []cardgamev1.Trait) []entity.Trait {
	result := make([]entity.Trait, 0, len(pbTraits))
	for _, pbTrait := range pbTraits {
		trait := TraitFromProto(pbTrait)
		if trait != "" {
			result = append(result, trait)
		}
	}
	return result
}

// GamePhaseFromProto GamePhase変換
func GamePhaseFromProto(pbPhase cardgamev1.GamePhase) entity.GamePhase {
	switch pbPhase {
	case cardgamev1.GamePhase_GAME_PHASE_TURN_START:
		return entity.GamePhaseTurnStart
	case cardgamev1.GamePhase_GAME_PHASE_DRAW:
		return entity.GamePhaseDraw
	case cardgamev1.GamePhase_GAME_PHASE_RESOURCE_GAIN:
		return entity.GamePhaseResourceGain
	case cardgamev1.GamePhase_GAME_PHASE_MAIN:
		return entity.GamePhaseMain
	case cardgamev1.GamePhase_GAME_PHASE_TURN_END:
		return entity.GamePhaseTurnEnd
	default:
		return entity.GamePhaseMain // デフォルト
	}
}

// ========================================
// ヘルパー関数
// ========================================

// optionalInt32ToPtrInt optional int32 を *int に変換
func optionalInt32ToPtrInt(v *int32) *int {
	if v == nil {
		return nil
	}
	val := int(*v)
	return &val
}

// ========================================
// CardEffect変換
// ========================================

// CardEffectFromProto ProtoからCardEffectに変換
func CardEffectFromProto(pb *cardgamev1.CardEffect) *entity.CardEffect {
	if pb == nil {
		return nil
	}

	definitions := make([]*entity.EffectDefinition, len(pb.Definitions))
	for i, def := range pb.Definitions {
		definitions[i] = effectDefinitionFromProto(def)
	}

	return &entity.CardEffect{
		Definitions: definitions,
	}
}

func effectDefinitionFromProto(pb *cardgamev1.EffectDefinition) *entity.EffectDefinition {
	if pb == nil {
		return nil
	}

	return &entity.EffectDefinition{
		RequireTarget: pb.RequireTarget,
		Root:          effectChainNodeFromProto(pb.Root),
	}
}

func effectChainNodeFromProto(pb *cardgamev1.EffectChainNode) *entity.EffectChainNode {
	if pb == nil {
		return nil
	}

	nodeType := effectChainNodeTypeFromProto(pb.Type)
	node := &entity.EffectChainNode{
		Type: nodeType,
	}

	// AtomicEffectがある場合はSequentialNodeに格納
	atomicEffect := atomicEffectFromProto(pb.AtomicEffect)

	switch pb.Type {
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_THEN:
		node.Sequential = &entity.SequentialNode{
			Effect: atomicEffect,
			Next:   effectChainNodeFromProto(pb.Next),
		}
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_AND:
		children := make([]*entity.EffectChainNode, len(pb.Children))
		for i, child := range pb.Children {
			children[i] = effectChainNodeFromProto(child)
		}
		node.Parallel = &entity.ParallelNode{
			Children: children,
		}
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_IF_ELSE:
		node.IfElse = &entity.IfElseNode{
			Condition: conditionFromProto(pb.Condition),
			Then:      effectChainNodeFromProto(pb.ThenNode),
			Else:      effectChainNodeFromProto(pb.ElseNode),
		}
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_REPEAT:
		var count int
		if pb.RepeatCount != nil {
			count = int(*pb.RepeatCount)
		}
		node.Repeat = &entity.RepeatNode{
			Effect: effectChainNodeFromProto(pb.RepeatEffect),
			Count:  count,
		}
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_FOREACH:
		var target entity.TargetSelector
		if pb.ForeachTarget != nil {
			if ts := targetSelectorFromProto(pb.ForeachTarget); ts != nil {
				target = *ts
			}
		}
		node.ForEach = &entity.ForEachNode{
			Target: target,
			Effect: effectChainNodeFromProto(pb.ForeachEffect),
		}
	}

	return node
}

func effectChainNodeTypeFromProto(pbType cardgamev1.EffectChainNodeType) entity.EffectOperator {
	switch pbType {
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_THEN:
		return entity.OperatorSequential
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_AND:
		return entity.OperatorParallel
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_IF_ELSE:
		return entity.OperatorIfElse
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_REPEAT:
		return entity.OperatorRepeat
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_FOREACH:
		return entity.OperatorForEach
	default:
		return entity.OperatorSequential
	}
}

func atomicEffectFromProto(pb *cardgamev1.AtomicEffect) *entity.AtomicEffect {
	if pb == nil {
		return nil
	}

	var target entity.TargetSelector
	if pb.Target != nil {
		if ts := targetSelectorFromProto(pb.Target); ts != nil {
			target = *ts
		}
	}

	effect := &entity.AtomicEffect{
		Type:       atomicEffectTypeFromProto(pb.Type),
		Target:     target,
		Timing:     effectTimingFromProto(pb.Timing),
		Parameters: make(map[string]any),
	}

	if pb.Value != nil {
		effect.Value = int(*pb.Value)
		effect.Parameters["amount"] = int(*pb.Value)
	}
	if pb.CardId != nil {
		effect.Parameters["card_id"] = *pb.CardId
	}
	if pb.Trait != nil {
		effect.Parameters["trait"] = TraitFromProto(*pb.Trait)
	}

	return effect
}

func atomicEffectTypeFromProto(pbType cardgamev1.AtomicEffectType) entity.AtomicEffectType {
	switch pbType {
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_DAMAGE:
		return entity.AtomicEffectDealDamage
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_SPLASH:
		return entity.AtomicEffectDealSplash
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_HP:
		return entity.AtomicEffectRestoreHP
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_MANA:
		return entity.AtomicEffectRestoreMana
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_FULL_RESTORE:
		return entity.AtomicEffectFullRestore
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DRAW_CARD:
		return entity.AtomicEffectDrawCard
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISCARD_CARD:
		return entity.AtomicEffectDiscardCard
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SEARCH_CARD:
		return entity.AtomicEffectSearchCard
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SHUFFLE_DECK:
		return entity.AtomicEffectShuffleDeck
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_ATTACK:
		return entity.AtomicEffectModifyAttack
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_DEFENSE:
		return entity.AtomicEffectModifyDefense
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_COST:
		return entity.AtomicEffectModifyCost
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_MAX_HP:
		return entity.AtomicEffectModifyMaxHP
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SUMMON_UNIT:
		return entity.AtomicEffectSummonUnit
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DESTROY_UNIT:
		return entity.AtomicEffectDestroyUnit
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_HAND:
		return entity.AtomicEffectReturnToHand
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_DECK:
		return entity.AtomicEffectReturnToDeck
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISABLE_UNIT:
		return entity.AtomicEffectSilenceUnit
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GRANT_TRAIT:
		return entity.AtomicEffectGrantTrait
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REMOVE_TRAIT:
		return entity.AtomicEffectRemoveTrait
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GAIN_MANA:
		return entity.AtomicEffectGainMana
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REDUCE_COST:
		return entity.AtomicEffectReduceCost
	default:
		return entity.AtomicEffectDealDamage
	}
}

func effectTimingFromProto(pbTiming cardgamev1.EffectTiming) entity.EffectTiming {
	switch pbTiming {
	case cardgamev1.EffectTiming_EFFECT_TIMING_IMMEDIATE:
		return entity.EffectTimingImmediate
	case cardgamev1.EffectTiming_EFFECT_TIMING_ON_SUMMON:
		return entity.EffectTimingOnSummon
	case cardgamev1.EffectTiming_EFFECT_TIMING_ON_DESTROY:
		return entity.EffectTimingOnDestroy
	case cardgamev1.EffectTiming_EFFECT_TIMING_ON_ATTACK:
		return entity.EffectTimingOnAttack
	case cardgamev1.EffectTiming_EFFECT_TIMING_ON_DAMAGED:
		return entity.EffectTimingOnDamaged
	case cardgamev1.EffectTiming_EFFECT_TIMING_TURN_START:
		return entity.EffectTimingTurnStart
	case cardgamev1.EffectTiming_EFFECT_TIMING_TURN_END:
		return entity.EffectTimingTurnEnd
	default:
		return entity.EffectTimingImmediate
	}
}

func targetSelectorFromProto(pb *cardgamev1.TargetSelector) *entity.TargetSelector {
	if pb == nil {
		return nil
	}

	return &entity.TargetSelector{
		Type:   targetTypeFromProto(pb.Type),
		Filter: conditionFilterFromProto(pb.Filter),
	}
}

func targetTypeFromProto(pbType cardgamev1.TargetType) entity.EffectTarget {
	switch pbType {
	case cardgamev1.TargetType_TARGET_TYPE_SELF:
		return entity.EffectTargetSelf
	case cardgamev1.TargetType_TARGET_TYPE_ENEMY_LEADER:
		return entity.EffectTargetOpponent
	case cardgamev1.TargetType_TARGET_TYPE_ALLY_LEADER:
		return entity.EffectTargetSelf
	case cardgamev1.TargetType_TARGET_TYPE_ENEMY_UNIT:
		return entity.EffectTargetEnemies
	case cardgamev1.TargetType_TARGET_TYPE_ALLY_UNIT:
		return entity.EffectTargetAllies
	case cardgamev1.TargetType_TARGET_TYPE_ALL_UNITS:
		return entity.EffectTargetAllUnits
	case cardgamev1.TargetType_TARGET_TYPE_ALL_ENEMY_UNITS:
		return entity.EffectTargetEnemies
	case cardgamev1.TargetType_TARGET_TYPE_ALL_ALLY_UNITS:
		return entity.EffectTargetAllies
	case cardgamev1.TargetType_TARGET_TYPE_RANDOM_ENEMY_UNIT:
		return entity.EffectTargetRandomEnemy
	case cardgamev1.TargetType_TARGET_TYPE_RANDOM_ALLY_UNIT:
		return entity.EffectTargetRandomAlly
	default:
		return entity.EffectTargetSpecific
	}
}

func conditionFromProto(pb *cardgamev1.ConditionFilter) *entity.Condition {
	if pb == nil {
		return nil
	}

	// TODO: ConditionFilterからConditionへの適切な変換ロジックを実装
	// 現状は簡易的な実装
	return &entity.Condition{
		Type:     entity.ConditionPlayerHP,
		Operator: entity.OperatorEqual,
		Value:    0,
	}
}

func conditionFilterFromProto(pb *cardgamev1.ConditionFilter) *entity.TargetFilter {
	if pb == nil {
		return nil
	}

	filter := &entity.TargetFilter{}

	// conditionTypeに基づいてフィルタを設定
	switch pb.ConditionType {
	case "HAS_TRAIT":
		if len(pb.Parameters) > 0 {
			// TODO: 文字列からTraitへの変換を実装
			filter.HasTrait = []entity.Trait{}
		}
	case "ATTACK_RANGE":
		if len(pb.Parameters) >= 2 {
			filter.MinAttack = intPtrFromString(pb.Parameters[0])
			filter.MaxAttack = intPtrFromString(pb.Parameters[1])
		}
	case "DEFENSE_RANGE":
		if len(pb.Parameters) >= 2 {
			filter.MinDefense = intPtrFromString(pb.Parameters[0])
			filter.MaxDefense = intPtrFromString(pb.Parameters[1])
		}
	}

	return filter
}

func stringPtrFromString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtrFromString(s string) *int {
	if s == "" {
		return nil
	}
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return nil
	}
	return &val
}

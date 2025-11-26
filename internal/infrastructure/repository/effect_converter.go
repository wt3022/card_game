package repository

import (
	"encoding/json"
	"fmt"

	cardgamev1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/internal/infrastructure/persistence/model"
)

// ========================================
// CardEffect変換（Proto → Model）
// Proto構造をDB永続化用のModel構造に変換
// クリーンアーキテクチャの依存関係を守るため、infrastructureレイヤーに配置
// ========================================

// CardEffectFromProtoToModel ProtoのCardEffectをModelに変換（interface{}版）
func CardEffectFromProtoToModel(pbInterface interface{}) (*model.CardEffectModel, error) {
	// 型アサーション
	pb, ok := pbInterface.(*cardgamev1.CardEffect)
	if !ok {
		return nil, fmt.Errorf("invalid type: expected *cardgamev1.CardEffect")
	}

	return cardEffectFromProtoToModelInternal(pb)
}

// cardEffectFromProtoToModelInternal 内部実装
func cardEffectFromProtoToModelInternal(pb *cardgamev1.CardEffect) (*model.CardEffectModel, error) {
	if pb == nil {
		return nil, nil
	}

	cardEffectModel := &model.CardEffectModel{
		CardID:      pb.CardId,
		Definitions: make([]model.EffectDefinitionModel, 0, len(pb.Definitions)),
	}

	for _, pbDef := range pb.Definitions {
		defModel, err := EffectDefinitionFromProtoToModel(pbDef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert effect definition: %w", err)
		}
		cardEffectModel.Definitions = append(cardEffectModel.Definitions, *defModel)
	}

	return cardEffectModel, nil
}

// EffectDefinitionFromProtoToModel ProtoのEffectDefinitionをModelに変換
func EffectDefinitionFromProtoToModel(pb *cardgamev1.EffectDefinition) (*model.EffectDefinitionModel, error) {
	if pb == nil {
		return nil, fmt.Errorf("effect definition cannot be nil")
	}

	defModel := &model.EffectDefinitionModel{
		RequireTarget: pb.RequireTarget,
	}

	if pb.Root != nil {
		rootModel, err := EffectChainNodeFromProtoToModel(pb.Root)
		if err != nil {
			return nil, fmt.Errorf("failed to convert root node: %w", err)
		}
		defModel.Root = rootModel
	}

	return defModel, nil
}

// EffectChainNodeFromProtoToModel ProtoのEffectChainNodeをModelに変換
func EffectChainNodeFromProtoToModel(pb *cardgamev1.EffectChainNode) (*model.EffectChainNodeModel, error) {
	if pb == nil {
		return nil, fmt.Errorf("effect chain node cannot be nil")
	}

	nodeModel := &model.EffectChainNodeModel{
		Type: EffectChainNodeTypeFromProtoToString(pb.Type),
	}

	// AtomicEffectを変換
	if pb.AtomicEffect != nil {
		atomicEffectModel, err := AtomicEffectFromProtoToModel(pb.AtomicEffect)
		if err != nil {
			return nil, fmt.Errorf("failed to convert atomic effect: %w", err)
		}
		nodeModel.AtomicEffect = atomicEffectModel
	}

	// ノードタイプに応じた具体構造を変換
	switch pb.Type {
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_THEN:
		if pb.Next != nil {
			nextModel, err := EffectChainNodeFromProtoToModel(pb.Next)
			if err != nil {
				return nil, fmt.Errorf("failed to convert next node: %w", err)
			}
			nodeModel.Sequential = &model.SequentialNodeModel{
				Next: nextModel,
			}
		} else {
			nodeModel.Sequential = &model.SequentialNodeModel{}
		}

	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_AND:
		parallelModel := &model.ParallelNodeModel{
			Children: make([]*model.EffectChainNodeModel, 0, len(pb.Children)),
		}
		for _, childPb := range pb.Children {
			childModel, err := EffectChainNodeFromProtoToModel(childPb)
			if err != nil {
				return nil, fmt.Errorf("failed to convert parallel child: %w", err)
			}
			parallelModel.Children = append(parallelModel.Children, childModel)
		}
		nodeModel.Parallel = parallelModel

	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_IF_ELSE:
		if pb.ThenNode == nil {
			return nil, fmt.Errorf("if_else node requires then_node")
		}
		if pb.Condition == nil {
			return nil, fmt.Errorf("if_else node requires condition")
		}

		thenModel, err := EffectChainNodeFromProtoToModel(pb.ThenNode)
		if err != nil {
			return nil, fmt.Errorf("failed to convert then node: %w", err)
		}

		conditionModel, err := ConditionFromProtoToModel(pb.Condition)
		if err != nil {
			return nil, fmt.Errorf("failed to convert condition: %w", err)
		}

		ifElseModel := &model.IfElseNodeModel{
			Then:      thenModel,
			Condition: conditionModel,
		}

		if pb.ElseNode != nil {
			elseModel, err := EffectChainNodeFromProtoToModel(pb.ElseNode)
			if err != nil {
				return nil, fmt.Errorf("failed to convert else node: %w", err)
			}
			ifElseModel.Else = elseModel
		}

		nodeModel.IfElse = ifElseModel

	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_REPEAT:
		if pb.RepeatEffect == nil {
			return nil, fmt.Errorf("repeat node requires repeat_effect")
		}

		repeatEffectModel, err := EffectChainNodeFromProtoToModel(pb.RepeatEffect)
		if err != nil {
			return nil, fmt.Errorf("failed to convert repeat effect: %w", err)
		}

		repeatCount := 1
		if pb.RepeatCount != nil {
			repeatCount = int(*pb.RepeatCount)
		}
		nodeModel.Repeat = &model.RepeatNodeModel{
			RepeatEffect: repeatEffectModel,
			RepeatCount:  repeatCount,
		}

	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_FOREACH:
		if pb.ForeachEffect == nil {
			return nil, fmt.Errorf("foreach node requires foreach_effect")
		}
		if pb.ForeachTarget == nil {
			return nil, fmt.Errorf("foreach node requires foreach_target")
		}

		foreachEffectModel, err := EffectChainNodeFromProtoToModel(pb.ForeachEffect)
		if err != nil {
			return nil, fmt.Errorf("failed to convert foreach effect: %w", err)
		}

		foreachTargetModel, err := TargetSelectorFromProtoToModel(pb.ForeachTarget)
		if err != nil {
			return nil, fmt.Errorf("failed to convert foreach target: %w", err)
		}

		nodeModel.ForEach = &model.ForEachNodeModel{
			ForEachEffect: foreachEffectModel,
			ForEachTarget: foreachTargetModel,
		}
	}

	return nodeModel, nil
}

// AtomicEffectFromProtoToModel ProtoのAtomicEffectをModelに変換
func AtomicEffectFromProtoToModel(pb *cardgamev1.AtomicEffect) (*model.AtomicEffectModel, error) {
	if pb == nil {
		return nil, fmt.Errorf("atomic effect cannot be nil")
	}

	targetModel, err := TargetSelectorFromProtoToModel(pb.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target selector: %w", err)
	}

	// Parametersは空のJSONオブジェクトとして初期化
	parametersJSON := "{}"

	value := 0
	if pb.Value != nil {
		value = int(*pb.Value)
	}

	atomicEffectModel := &model.AtomicEffectModel{
		Type:       AtomicEffectTypeFromProtoToString(pb.Type),
		Value:      value,
		Timing:     "Immediate", // Protoではタイミングは常にImmediate
		Target:     targetModel,
		Parameters: parametersJSON,
		Multiplier: 1.0,
	}

	// CardIDやTraitをParametersに格納（必要に応じて）
	if (pb.CardId != nil && *pb.CardId != "") || (pb.Trait != nil && *pb.Trait != cardgamev1.Trait_TRAIT_UNSPECIFIED) {
		params := make(map[string]interface{})
		if pb.CardId != nil && *pb.CardId != "" {
			params["card_id"] = *pb.CardId
		}
		if pb.Trait != nil && *pb.Trait != cardgamev1.Trait_TRAIT_UNSPECIFIED {
			params["trait"] = TraitFromProto(*pb.Trait)
		}
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal parameters: %w", err)
		}
		atomicEffectModel.Parameters = string(paramsBytes)
	}

	return atomicEffectModel, nil
}

// TargetSelectorFromProtoToModel ProtoのTargetSelectorをModelに変換
func TargetSelectorFromProtoToModel(pb *cardgamev1.TargetSelector) (*model.TargetSelectorModel, error) {
	if pb == nil {
		return nil, fmt.Errorf("target selector cannot be nil")
	}

	targetModel := &model.TargetSelectorModel{
		Type: TargetTypeFromProtoToString(pb.Type),
	}

	if pb.Filter != nil {
		filterModel, err := ConditionFilterFromProtoToModel(pb.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter: %w", err)
		}
		targetModel.Filter = filterModel
	}

	return targetModel, nil
}

// ConditionFilterFromProtoToModel ProtoのConditionFilterをTargetFilterModelに変換
func ConditionFilterFromProtoToModel(pb *cardgamev1.ConditionFilter) (*model.TargetFilterModel, error) {
	if pb == nil {
		return nil, fmt.Errorf("condition filter cannot be nil")
	}

	filterModel := &model.TargetFilterModel{}

	// ConditionFilterのパラメータを解析してTargetFilterModelに変換
	// condition_typeとparametersに基づいて適切なフィールドを設定
	switch pb.ConditionType {
	case "HAS_TRAIT":
		if len(pb.Parameters) > 0 {
			for _, param := range pb.Parameters {
				filterModel.HasTraits = append(filterModel.HasTraits, model.TargetFilterTraitModel{
					Trait:      param,
					IsHasTrait: true,
				})
			}
		}
	case "ATTACK_RANGE":
		if len(pb.Parameters) >= 2 {
			minAttack := parseIntPtr(pb.Parameters[0])
			maxAttack := parseIntPtr(pb.Parameters[1])
			filterModel.MinAttack = minAttack
			filterModel.MaxAttack = maxAttack
		}
	case "DEFENSE_RANGE":
		if len(pb.Parameters) >= 2 {
			minDefense := parseIntPtr(pb.Parameters[0])
			maxDefense := parseIntPtr(pb.Parameters[1])
			filterModel.MinDefense = minDefense
			filterModel.MaxDefense = maxDefense
		}
	case "COST_RANGE":
		if len(pb.Parameters) >= 2 {
			minCost := parseIntPtr(pb.Parameters[0])
			maxCost := parseIntPtr(pb.Parameters[1])
			filterModel.MinCost = minCost
			filterModel.MaxCost = maxCost
		}
	case "CARD_TYPE":
		if len(pb.Parameters) > 0 {
			filterModel.CardType = pb.Parameters[0]
		}
	}

	return filterModel, nil
}

// ConditionFromProtoToModel ProtoのConditionFilterをConditionModelに変換（IF_ELSE用）
func ConditionFromProtoToModel(pb *cardgamev1.ConditionFilter) (*model.ConditionModel, error) {
	if pb == nil {
		return nil, fmt.Errorf("condition cannot be nil")
	}

	conditionModel := &model.ConditionModel{
		Type: pb.ConditionType,
	}

	// Operatorとvalueをparametersから抽出
	if len(pb.Parameters) >= 2 {
		conditionModel.Operator = pb.Parameters[0]
		if val := parseIntPtr(pb.Parameters[1]); val != nil {
			conditionModel.Value = *val
		}
	}

	return conditionModel, nil
}

// ========================================
// Enum変換ヘルパー（Proto → String）
// ========================================

// EffectChainNodeTypeFromProtoToString EffectChainNodeTypeをStringに変換
func EffectChainNodeTypeFromProtoToString(pbType cardgamev1.EffectChainNodeType) string {
	switch pbType {
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_THEN:
		return "THEN"
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_AND:
		return "AND"
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_IF_ELSE:
		return "IF_ELSE"
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_REPEAT:
		return "REPEAT"
	case cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_FOREACH:
		return "FOREACH"
	default:
		return "THEN"
	}
}

// AtomicEffectTypeFromProtoToString AtomicEffectTypeをStringに変換
func AtomicEffectTypeFromProtoToString(pbType cardgamev1.AtomicEffectType) string {
	switch pbType {
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_DAMAGE:
		return "DEAL_DAMAGE"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_SPLASH:
		return "DEAL_SPLASH"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_HP:
		return "RESTORE_HP"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_MANA:
		return "RESTORE_MANA"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_FULL_RESTORE:
		return "FULL_RESTORE"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DRAW_CARD:
		return "DRAW_CARD"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISCARD_CARD:
		return "DISCARD_CARD"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SEARCH_CARD:
		return "SEARCH_CARD"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SHUFFLE_DECK:
		return "SHUFFLE_DECK"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_ATTACK:
		return "MODIFY_ATTACK"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_DEFENSE:
		return "MODIFY_DEFENSE"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_COST:
		return "MODIFY_COST"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_MAX_HP:
		return "MODIFY_MAX_HP"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SUMMON_UNIT:
		return "SUMMON_UNIT"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DESTROY_UNIT:
		return "DESTROY_UNIT"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_HAND:
		return "RETURN_TO_HAND"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_DECK:
		return "RETURN_TO_DECK"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISABLE_UNIT:
		return "DISABLE_UNIT"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GRANT_TRAIT:
		return "GRANT_TRAIT"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REMOVE_TRAIT:
		return "REMOVE_TRAIT"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GAIN_MANA:
		return "GAIN_MANA"
	case cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REDUCE_COST:
		return "REDUCE_COST"
	default:
		return "DEAL_DAMAGE"
	}
}

// TargetTypeFromProtoToString TargetTypeをStringに変換
func TargetTypeFromProtoToString(pbType cardgamev1.TargetType) string {
	switch pbType {
	case cardgamev1.TargetType_TARGET_TYPE_SELF:
		return "Self"
	case cardgamev1.TargetType_TARGET_TYPE_ENEMY_LEADER:
		return "Opponent"
	case cardgamev1.TargetType_TARGET_TYPE_ALLY_LEADER:
		return "Self"
	case cardgamev1.TargetType_TARGET_TYPE_ENEMY_UNIT:
		return "Enemies"
	case cardgamev1.TargetType_TARGET_TYPE_ALLY_UNIT:
		return "Allies"
	case cardgamev1.TargetType_TARGET_TYPE_ALL_UNITS:
		return "AllUnits"
	case cardgamev1.TargetType_TARGET_TYPE_ALL_ENEMY_UNITS:
		return "Enemies"
	case cardgamev1.TargetType_TARGET_TYPE_ALL_ALLY_UNITS:
		return "Allies"
	case cardgamev1.TargetType_TARGET_TYPE_RANDOM_ENEMY_UNIT:
		return "RandomEnemy"
	case cardgamev1.TargetType_TARGET_TYPE_RANDOM_ALLY_UNIT:
		return "RandomAlly"
	default:
		return "Specific"
	}
}

// TraitFromProto Trait変換（repository層専用、string形式に変換）
// Note: adapter/converter層にも同名の関数があるが、こちらはstring形式に変換する点が異なる
func TraitFromProto(pbTrait cardgamev1.Trait) string {
	switch pbTrait {
	case cardgamev1.Trait_TRAIT_RUSH:
		return "RUSH"
	case cardgamev1.Trait_TRAIT_CHARGE:
		return "CHARGE"
	case cardgamev1.Trait_TRAIT_WINDFURY:
		return "WINDFURY"
	case cardgamev1.Trait_TRAIT_PIERCE:
		return "PIERCE"
	case cardgamev1.Trait_TRAIT_GUARDIAN:
		return "GUARDIAN"
	case cardgamev1.Trait_TRAIT_EFFECT_SHIELD:
		return "EFFECT_SHIELD"
	case cardgamev1.Trait_TRAIT_UNTARGETABLE:
		return "UNTARGETABLE"
	default:
		return ""
	}
}

// parseIntPtr 文字列をintポインタに変換
func parseIntPtr(s string) *int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err == nil {
		return &val
	}
	return nil
}

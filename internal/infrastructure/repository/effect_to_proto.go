package repository

import (
	"encoding/json"
	"fmt"

	cardgamev1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/internal/infrastructure/persistence/model"
)

// ========================================
// CardEffect変換（Model → Proto）
// DB永続化用のModel構造をProto構造に変換
// クリーンアーキテクチャの依存関係を守るため、infrastructureレイヤーに配置
// ========================================

// CardEffectModelToProto ModelのCardEffectをProtoに変換
func CardEffectModelToProto(cardEffectModel *model.CardEffectModel) (*cardgamev1.CardEffect, error) {
	if cardEffectModel == nil {
		return nil, nil
	}

	cardEffect := &cardgamev1.CardEffect{
		Id:          uint32(cardEffectModel.ID),
		CardId:      cardEffectModel.CardID,
		Definitions: make([]*cardgamev1.EffectDefinition, 0, len(cardEffectModel.Definitions)),
	}

	for _, defModel := range cardEffectModel.Definitions {
		defProto, err := EffectDefinitionModelToProto(&defModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert effect definition: %w", err)
		}
		cardEffect.Definitions = append(cardEffect.Definitions, defProto)
	}

	return cardEffect, nil
}

// EffectDefinitionModelToProto ModelのEffectDefinitionをProtoに変換
func EffectDefinitionModelToProto(defModel *model.EffectDefinitionModel) (*cardgamev1.EffectDefinition, error) {
	if defModel == nil {
		return nil, fmt.Errorf("effect definition model cannot be nil")
	}

	defProto := &cardgamev1.EffectDefinition{
		Id:            uint32(defModel.ID),
		RequireTarget: defModel.RequireTarget,
	}

	if defModel.Root != nil {
		rootProto, err := EffectChainNodeModelToProto(defModel.Root)
		if err != nil {
			return nil, fmt.Errorf("failed to convert root node: %w", err)
		}
		defProto.Root = rootProto
	}

	return defProto, nil
}

// EffectChainNodeModelToProto ModelのEffectChainNodeをProtoに変換
func EffectChainNodeModelToProto(nodeModel *model.EffectChainNodeModel) (*cardgamev1.EffectChainNode, error) {
	if nodeModel == nil {
		return nil, fmt.Errorf("effect chain node model cannot be nil")
	}

	nodeProto := &cardgamev1.EffectChainNode{
		Id:   uint32(nodeModel.ID),
		Type: EffectChainNodeTypeStringToProto(nodeModel.Type),
	}

	// AtomicEffectを変換
	if nodeModel.AtomicEffect != nil {
		atomicEffectProto, err := AtomicEffectModelToProto(nodeModel.AtomicEffect)
		if err != nil {
			return nil, fmt.Errorf("failed to convert atomic effect: %w", err)
		}
		nodeProto.AtomicEffect = atomicEffectProto
	}

	// ノードタイプに応じた具体構造を変換
	switch nodeModel.Type {
	case "THEN":
		if nodeModel.Sequential != nil && nodeModel.Sequential.Next != nil {
			nextProto, err := EffectChainNodeModelToProto(nodeModel.Sequential.Next)
			if err != nil {
				return nil, fmt.Errorf("failed to convert next node: %w", err)
			}
			nodeProto.Next = nextProto
		}

	case "AND":
		if nodeModel.Parallel != nil {
			for _, childModel := range nodeModel.Parallel.Children {
				childProto, err := EffectChainNodeModelToProto(childModel)
				if err != nil {
					return nil, fmt.Errorf("failed to convert parallel child: %w", err)
				}
				nodeProto.Children = append(nodeProto.Children, childProto)
			}
		}

	case "IF_ELSE":
		if nodeModel.IfElse != nil {
			if nodeModel.IfElse.Then == nil {
				return nil, fmt.Errorf("if_else node requires then node")
			}
			if nodeModel.IfElse.Condition == nil {
				return nil, fmt.Errorf("if_else node requires condition")
			}

			thenProto, err := EffectChainNodeModelToProto(nodeModel.IfElse.Then)
			if err != nil {
				return nil, fmt.Errorf("failed to convert then node: %w", err)
			}
			nodeProto.ThenNode = thenProto

			conditionProto, err := ConditionModelToProto(nodeModel.IfElse.Condition)
			if err != nil {
				return nil, fmt.Errorf("failed to convert condition: %w", err)
			}
			nodeProto.Condition = conditionProto

			if nodeModel.IfElse.Else != nil {
				elseProto, err := EffectChainNodeModelToProto(nodeModel.IfElse.Else)
				if err != nil {
					return nil, fmt.Errorf("failed to convert else node: %w", err)
				}
				nodeProto.ElseNode = elseProto
			}
		}

	case "REPEAT":
		if nodeModel.Repeat != nil {
			if nodeModel.Repeat.RepeatEffect == nil {
				return nil, fmt.Errorf("repeat node requires repeat effect")
			}

			repeatEffectProto, err := EffectChainNodeModelToProto(nodeModel.Repeat.RepeatEffect)
			if err != nil {
				return nil, fmt.Errorf("failed to convert repeat effect: %w", err)
			}
			nodeProto.RepeatEffect = repeatEffectProto
			repeatCount := int32(nodeModel.Repeat.RepeatCount)
			nodeProto.RepeatCount = &repeatCount
		}

	case "FOREACH":
		if nodeModel.ForEach != nil {
			if nodeModel.ForEach.ForEachEffect == nil {
				return nil, fmt.Errorf("foreach node requires foreach effect")
			}
			if nodeModel.ForEach.ForEachTarget == nil {
				return nil, fmt.Errorf("foreach node requires foreach target")
			}

			foreachEffectProto, err := EffectChainNodeModelToProto(nodeModel.ForEach.ForEachEffect)
			if err != nil {
				return nil, fmt.Errorf("failed to convert foreach effect: %w", err)
			}
			nodeProto.ForeachEffect = foreachEffectProto

			foreachTargetProto, err := TargetSelectorModelToProto(nodeModel.ForEach.ForEachTarget)
			if err != nil {
				return nil, fmt.Errorf("failed to convert foreach target: %w", err)
			}
			nodeProto.ForeachTarget = foreachTargetProto
		}
	}

	return nodeProto, nil
}

// AtomicEffectModelToProto ModelのAtomicEffectをProtoに変換
func AtomicEffectModelToProto(atomicEffectModel *model.AtomicEffectModel) (*cardgamev1.AtomicEffect, error) {
	if atomicEffectModel == nil {
		return nil, fmt.Errorf("atomic effect model cannot be nil")
	}

	targetProto, err := TargetSelectorModelToProto(atomicEffectModel.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target selector: %w", err)
	}

	value := int32(atomicEffectModel.Value)
	atomicEffectProto := &cardgamev1.AtomicEffect{
		Id:     uint32(atomicEffectModel.ID),
		Type:   AtomicEffectTypeStringToProto(atomicEffectModel.Type),
		Target: targetProto,
		Value:  &value,
	}

	// ParametersからCardIDやTraitを抽出
	if atomicEffectModel.Parameters != "" && atomicEffectModel.Parameters != "{}" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(atomicEffectModel.Parameters), &params); err == nil {
			if cardID, ok := params["card_id"].(string); ok {
				atomicEffectProto.CardId = &cardID
			}
			if traitStr, ok := params["trait"].(string); ok {
				trait := TraitStringToProto(traitStr)
				atomicEffectProto.Trait = &trait
			}
		}
	}

	return atomicEffectProto, nil
}

// TargetSelectorModelToProto ModelのTargetSelectorをProtoに変換
func TargetSelectorModelToProto(targetModel *model.TargetSelectorModel) (*cardgamev1.TargetSelector, error) {
	if targetModel == nil {
		return nil, fmt.Errorf("target selector model cannot be nil")
	}

	targetProto := &cardgamev1.TargetSelector{
		Id:   uint32(targetModel.ID),
		Type: TargetTypeStringToProto(targetModel.Type),
	}

	if targetModel.Filter != nil {
		filterProto, err := TargetFilterModelToProto(targetModel.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter: %w", err)
		}
		targetProto.Filter = filterProto
	}

	return targetProto, nil
}

// TargetFilterModelToProto ModelのTargetFilterをConditionFilterProtoに変換
func TargetFilterModelToProto(filterModel *model.TargetFilterModel) (*cardgamev1.ConditionFilter, error) {
	if filterModel == nil {
		return nil, fmt.Errorf("target filter model cannot be nil")
	}

	filterProto := &cardgamev1.ConditionFilter{
		Id:         uint32(filterModel.ID),
		Parameters: []string{},
	}

	// フィルタタイプに応じて変換
	if len(filterModel.HasTraits) > 0 {
		filterProto.ConditionType = "HAS_TRAIT"
		for _, traitModel := range filterModel.HasTraits {
			if traitModel.IsHasTrait {
				filterProto.Parameters = append(filterProto.Parameters, traitModel.Trait)
			}
		}
	} else if filterModel.MinAttack != nil || filterModel.MaxAttack != nil {
		filterProto.ConditionType = "ATTACK_RANGE"
		filterProto.Parameters = []string{
			intPtrToString(filterModel.MinAttack),
			intPtrToString(filterModel.MaxAttack),
		}
	} else if filterModel.MinDefense != nil || filterModel.MaxDefense != nil {
		filterProto.ConditionType = "DEFENSE_RANGE"
		filterProto.Parameters = []string{
			intPtrToString(filterModel.MinDefense),
			intPtrToString(filterModel.MaxDefense),
		}
	} else if filterModel.MinCost != nil || filterModel.MaxCost != nil {
		filterProto.ConditionType = "COST_RANGE"
		filterProto.Parameters = []string{
			intPtrToString(filterModel.MinCost),
			intPtrToString(filterModel.MaxCost),
		}
	} else if filterModel.CardType != "" {
		filterProto.ConditionType = "CARD_TYPE"
		filterProto.Parameters = []string{filterModel.CardType}
	}

	return filterProto, nil
}

// ConditionModelToProto ModelのConditionをConditionFilterProtoに変換
func ConditionModelToProto(conditionModel *model.ConditionModel) (*cardgamev1.ConditionFilter, error) {
	if conditionModel == nil {
		return nil, fmt.Errorf("condition model cannot be nil")
	}

	conditionProto := &cardgamev1.ConditionFilter{
		Id:            uint32(conditionModel.ID),
		ConditionType: conditionModel.Type,
		Parameters: []string{
			conditionModel.Operator,
			fmt.Sprintf("%d", conditionModel.Value),
		},
	}

	return conditionProto, nil
}

// ========================================
// Enum変換ヘルパー（String → Proto）
// ========================================

// EffectChainNodeTypeStringToProto StringをEffectChainNodeTypeに変換
func EffectChainNodeTypeStringToProto(nodeType string) cardgamev1.EffectChainNodeType {
	switch nodeType {
	case "THEN":
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_THEN
	case "AND":
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_AND
	case "IF_ELSE":
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_IF_ELSE
	case "REPEAT":
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_REPEAT
	case "FOREACH":
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_FOREACH
	default:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_UNSPECIFIED
	}
}

// AtomicEffectTypeStringToProto StringをAtomicEffectTypeに変換
func AtomicEffectTypeStringToProto(effectType string) cardgamev1.AtomicEffectType {
	switch effectType {
	case "DEAL_DAMAGE":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_DAMAGE
	case "DEAL_SPLASH":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_SPLASH
	case "RESTORE_HP":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_HP
	case "RESTORE_MANA":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_MANA
	case "FULL_RESTORE":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_FULL_RESTORE
	case "DRAW_CARD":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DRAW_CARD
	case "DISCARD_CARD":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISCARD_CARD
	case "SEARCH_CARD":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SEARCH_CARD
	case "SHUFFLE_DECK":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SHUFFLE_DECK
	case "MODIFY_ATTACK":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_ATTACK
	case "MODIFY_DEFENSE":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_DEFENSE
	case "MODIFY_COST":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_COST
	case "MODIFY_MAX_HP":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_MAX_HP
	case "SUMMON_UNIT":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SUMMON_UNIT
	case "DESTROY_UNIT":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DESTROY_UNIT
	case "RETURN_TO_HAND":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_HAND
	case "RETURN_TO_DECK":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_DECK
	case "DISABLE_UNIT":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISABLE_UNIT
	case "GRANT_TRAIT":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GRANT_TRAIT
	case "REMOVE_TRAIT":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REMOVE_TRAIT
	case "GAIN_MANA":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GAIN_MANA
	case "REDUCE_COST":
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REDUCE_COST
	default:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_UNSPECIFIED
	}
}

// TargetTypeStringToProto StringをTargetTypeに変換
func TargetTypeStringToProto(targetType string) cardgamev1.TargetType {
	switch targetType {
	case "Self":
		return cardgamev1.TargetType_TARGET_TYPE_SELF
	case "Opponent":
		return cardgamev1.TargetType_TARGET_TYPE_ENEMY_LEADER
	case "Allies":
		return cardgamev1.TargetType_TARGET_TYPE_ALL_ALLY_UNITS
	case "Enemies":
		return cardgamev1.TargetType_TARGET_TYPE_ALL_ENEMY_UNITS
	case "AllUnits":
		return cardgamev1.TargetType_TARGET_TYPE_ALL_UNITS
	case "RandomAlly":
		return cardgamev1.TargetType_TARGET_TYPE_RANDOM_ALLY_UNIT
	case "RandomEnemy":
		return cardgamev1.TargetType_TARGET_TYPE_RANDOM_ENEMY_UNIT
	case "Specific":
		return cardgamev1.TargetType_TARGET_TYPE_UNSPECIFIED
	default:
		return cardgamev1.TargetType_TARGET_TYPE_UNSPECIFIED
	}
}

// TraitStringToProto StringをTraitに変換
func TraitStringToProto(trait string) cardgamev1.Trait {
	switch trait {
	case "RUSH":
		return cardgamev1.Trait_TRAIT_RUSH
	case "CHARGE":
		return cardgamev1.Trait_TRAIT_CHARGE
	case "WINDFURY":
		return cardgamev1.Trait_TRAIT_WINDFURY
	case "PIERCE":
		return cardgamev1.Trait_TRAIT_PIERCE
	case "GUARDIAN":
		return cardgamev1.Trait_TRAIT_GUARDIAN
	case "EFFECT_SHIELD":
		return cardgamev1.Trait_TRAIT_EFFECT_SHIELD
	case "UNTARGETABLE":
		return cardgamev1.Trait_TRAIT_UNTARGETABLE
	default:
		return cardgamev1.Trait_TRAIT_UNSPECIFIED
	}
}

// intPtrToString intポインタを文字列に変換
func intPtrToString(v *int) string {
	if v == nil {
		return "0"
	}
	return fmt.Sprintf("%d", *v)
}

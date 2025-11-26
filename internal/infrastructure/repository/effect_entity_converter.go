package repository

import (
	"card_game/internal/core/entity"
	"card_game/internal/infrastructure/persistence/model"
)

// CardEffectFromEntityToModel entity.CardEffectをmodel.CardEffectModelに変換
func CardEffectFromEntityToModel(cardEffect *entity.CardEffect) (*model.CardEffectModel, error) {
	if cardEffect == nil {
		return nil, nil
	}

	cardEffectModel := &model.CardEffectModel{
		Definitions: make([]model.EffectDefinitionModel, len(cardEffect.Definitions)),
	}

	for i, def := range cardEffect.Definitions {
		cardEffectModel.Definitions[i] = effectDefinitionFromEntityToModel(def)
	}

	return cardEffectModel, nil
}

func effectDefinitionFromEntityToModel(def *entity.EffectDefinition) model.EffectDefinitionModel {
	if def == nil {
		return model.EffectDefinitionModel{}
	}

	return model.EffectDefinitionModel{
		RequireTarget: def.RequireTarget,
		Root:          effectChainNodeFromEntityToModel(def.Root),
	}
}

func effectChainNodeFromEntityToModel(node *entity.EffectChainNode) *model.EffectChainNodeModel {
	if node == nil {
		return nil
	}

	nodeModel := &model.EffectChainNodeModel{
		Type: string(node.Type),
	}

	switch node.Type {
	case entity.OperatorSequential:
		if node.Sequential != nil {
			nodeModel.AtomicEffect = atomicEffectFromEntityToModel(node.Sequential.Effect)
			nodeModel.Sequential = &model.SequentialNodeModel{
				Next: effectChainNodeFromEntityToModel(node.Sequential.Next),
			}
		}
	case entity.OperatorParallel:
		if node.Parallel != nil {
			children := make([]*model.EffectChainNodeModel, len(node.Parallel.Children))
			for i, child := range node.Parallel.Children {
				children[i] = effectChainNodeFromEntityToModel(child)
			}
			nodeModel.Parallel = &model.ParallelNodeModel{
				Children: children,
			}
		}
	case entity.OperatorIfElse:
		if node.IfElse != nil {
			nodeModel.IfElse = &model.IfElseNodeModel{
				Condition: conditionFromEntityToModel(node.IfElse.Condition),
				Then:      effectChainNodeFromEntityToModel(node.IfElse.Then),
				Else:      effectChainNodeFromEntityToModel(node.IfElse.Else),
			}
		}
	case entity.OperatorRepeat:
		if node.Repeat != nil {
			repeatEffect := effectChainNodeFromEntityToModel(node.Repeat.Effect)
			nodeModel.Repeat = &model.RepeatNodeModel{
				RepeatCount:  node.Repeat.Count,
				RepeatEffect: repeatEffect,
			}
		}
	case entity.OperatorForEach:
		if node.ForEach != nil {
			forEachEffect := effectChainNodeFromEntityToModel(node.ForEach.Effect)
			forEachTarget := targetSelectorFromEntityToModel(&node.ForEach.Target)
			nodeModel.ForEach = &model.ForEachNodeModel{
				ForEachTarget: forEachTarget,
				ForEachEffect: forEachEffect,
			}
		}
	}

	return nodeModel
}

func atomicEffectFromEntityToModel(effect *entity.AtomicEffect) *model.AtomicEffectModel {
	if effect == nil {
		return nil
	}

	effectModel := &model.AtomicEffectModel{
		Type:   string(effect.Type),
		Target: targetSelectorFromEntityToModel(&effect.Target),
		Value:  effect.Value,
		Timing: string(effect.Timing),
	}

	// ParametersをJSON文字列として保存
	if len(effect.Parameters) > 0 {
		// 簡易的にJSON文字列化（実装は省略）
		effectModel.Parameters = "{}"
	}

	return effectModel
}

func targetSelectorFromEntityToModel(selector *entity.TargetSelector) *model.TargetSelectorModel {
	if selector == nil {
		return nil
	}

	return &model.TargetSelectorModel{
		Type:   string(selector.Type),
		Filter: targetFilterFromEntityToModel(selector.Filter),
	}
}

func targetFilterFromEntityToModel(filter *entity.TargetFilter) *model.TargetFilterModel {
	if filter == nil {
		return nil
	}

	// 簡易的な変換
	return &model.TargetFilterModel{}
}

func conditionFromEntityToModel(cond *entity.Condition) *model.ConditionModel {
	if cond == nil {
		return nil
	}

	return &model.ConditionModel{
		Type:     string(cond.Type),
		Operator: string(cond.Operator),
		Value:    cond.Value,
	}
}

// CardEffectFromModelToEntity model.CardEffectModelをentity.CardEffectに変換
func CardEffectFromModelToEntity(cardEffectModel *model.CardEffectModel) (*entity.CardEffect, error) {
	if cardEffectModel == nil {
		return nil, nil
	}

	cardEffect := &entity.CardEffect{
		Definitions: make([]*entity.EffectDefinition, len(cardEffectModel.Definitions)),
	}

	for i, def := range cardEffectModel.Definitions {
		cardEffect.Definitions[i] = effectDefinitionFromModelToEntity(&def)
	}

	return cardEffect, nil
}

func effectDefinitionFromModelToEntity(def *model.EffectDefinitionModel) *entity.EffectDefinition {
	if def == nil {
		return nil
	}

	return &entity.EffectDefinition{
		RequireTarget: def.RequireTarget,
		Root:          effectChainNodeFromModelToEntity(def.Root),
	}
}

func effectChainNodeFromModelToEntity(node *model.EffectChainNodeModel) *entity.EffectChainNode {
	if node == nil {
		return nil
	}

	entityNode := &entity.EffectChainNode{
		Type: entity.EffectOperator(node.Type),
	}

	switch node.Type {
	case "THEN":
		if node.Sequential != nil || node.AtomicEffect != nil {
			entityNode.Sequential = &entity.SequentialNode{
				Effect: atomicEffectFromModelToEntity(node.AtomicEffect),
				Next:   effectChainNodeFromModelToEntity(node.Sequential.Next),
			}
		}
	case "AND":
		if node.Parallel != nil {
			children := make([]*entity.EffectChainNode, len(node.Parallel.Children))
			for i, child := range node.Parallel.Children {
				children[i] = effectChainNodeFromModelToEntity(child)
			}
			entityNode.Parallel = &entity.ParallelNode{
				Children: children,
			}
		}
	case "IF_ELSE":
		if node.IfElse != nil {
			entityNode.IfElse = &entity.IfElseNode{
				Condition: conditionFromModelToEntity(node.IfElse.Condition),
				Then:      effectChainNodeFromModelToEntity(node.IfElse.Then),
				Else:      effectChainNodeFromModelToEntity(node.IfElse.Else),
			}
		}
	case "REPEAT":
		if node.Repeat != nil {
			entityNode.Repeat = &entity.RepeatNode{
				Count:  node.Repeat.RepeatCount,
				Effect: effectChainNodeFromModelToEntity(node.Repeat.RepeatEffect),
			}
		}
	case "FOREACH":
		if node.ForEach != nil {
			var target entity.TargetSelector
			if node.ForEach.ForEachTarget != nil {
				if ts := targetSelectorFromModelToEntity(node.ForEach.ForEachTarget); ts != nil {
					target = *ts
				}
			}
			entityNode.ForEach = &entity.ForEachNode{
				Target: target,
				Effect: effectChainNodeFromModelToEntity(node.ForEach.ForEachEffect),
			}
		}
	}

	return entityNode
}

func atomicEffectFromModelToEntity(effectModel *model.AtomicEffectModel) *entity.AtomicEffect {
	if effectModel == nil {
		return nil
	}

	effect := &entity.AtomicEffect{
		Type:       entity.AtomicEffectType(effectModel.Type),
		Target:     *targetSelectorFromModelToEntity(effectModel.Target),
		Value:      effectModel.Value,
		Timing:     entity.EffectTimingImmediate,
		Parameters: make(map[string]any),
	}

	// ParametersはJSON文字列からデシリアライズ（簡易実装）
	if effectModel.Parameters != "" {
		// TODO: JSONデシリアライズ実装
	}

	return effect
}

func targetSelectorFromModelToEntity(selectorModel *model.TargetSelectorModel) *entity.TargetSelector {
	if selectorModel == nil {
		return &entity.TargetSelector{
			Type: entity.EffectTargetSelf,
		}
	}

	return &entity.TargetSelector{
		Type:   entity.EffectTarget(selectorModel.Type),
		Filter: targetFilterFromModelToEntity(selectorModel.Filter),
	}
}

func targetFilterFromModelToEntity(filterModel *model.TargetFilterModel) *entity.TargetFilter {
	if filterModel == nil {
		return nil
	}

	// 簡易的な変換
	return &entity.TargetFilter{}
}

func conditionFromModelToEntity(condModel *model.ConditionModel) *entity.Condition {
	if condModel == nil {
		return nil
	}

	return &entity.Condition{
		Type:     entity.ConditionType(condModel.Type),
		Operator: entity.ComparisonOperator(condModel.Operator),
		Value:    condModel.Value,
	}
}

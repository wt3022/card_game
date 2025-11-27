package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"card_game/internal/infrastructure/persistence/model"
)

// GenerateEffectDescription CardEffectModelから人間が読める効果テキストを生成
func GenerateEffectDescription(cardEffectModel *model.CardEffectModel) string {
	if cardEffectModel == nil || len(cardEffectModel.Definitions) == 0 {
		return ""
	}

	var descriptions []string
	seen := make(map[string]bool) // 重複チェック用

	for _, def := range cardEffectModel.Definitions {
		desc := generateDefinitionDescription(&def)
		if desc != "" && !seen[desc] {
			descriptions = append(descriptions, desc)
			seen[desc] = true
		}
	}

	return strings.Join(descriptions, " ")
}

// generateDefinitionDescription 単一のEffectDefinitionから説明文を生成
func generateDefinitionDescription(def *model.EffectDefinitionModel) string {
	if def.Root == nil {
		return ""
	}

	return generateNodeDescription(def.Root)
}

// generateNodeDescription EffectChainNodeから説明文を生成
func generateNodeDescription(node *model.EffectChainNodeModel) string {
	if node == nil {
		return ""
	}

	// AtomicEffectがある場合は、Typeに関わらずAtomicノードとして処理
	if node.AtomicEffect != nil {
		desc := generateAtomicEffectDescription(node.AtomicEffect)

		// THENタイプの場合、次のノードも処理して連結
		if node.Type == "THEN" && node.Sequential != nil && node.Sequential.Next != nil {
			nextDesc := generateNodeDescription(node.Sequential.Next)
			if nextDesc != "" {
				if desc != "" {
					return desc + "、" + nextDesc
				}
				return nextDesc
			}
		}

		return desc
	}

	// AtomicEffectがない場合は、Typeに応じて処理
	switch node.Type {
	case "THEN":
		return generateSequenceDescription(node)
	case "AND":
		return generateChoiceDescription(node)
	case "IF_ELSE":
		return generateIfElseDescription(node)
	case "REPEAT":
		return generateRepeatDescription(node)
	case "FOREACH":
		return generateForEachDescription(node)
	default:
		return ""
	}
}

// generateAtomicEffectDescription AtomicEffectから説明文を生成
func generateAtomicEffectDescription(atomicEffect *model.AtomicEffectModel) string {
	if atomicEffect == nil {
		return ""
	}

	switch atomicEffect.Type {
	case "DEAL_DAMAGE":
		return generateDamageDescription(atomicEffect)
	case "DEAL_SPLASH":
		return generateSplashDamageDescription(atomicEffect)
	case "RESTORE_HP":
		return generateHealDescription(atomicEffect)
	case "RESTORE_MANA":
		return generateRestoreManaDescription(atomicEffect)
	case "FULL_RESTORE":
		return "完全回復"
	case "DRAW_CARD":
		return generateDrawCardDescription(atomicEffect)
	case "DISCARD_CARD":
		return generateDiscardDescription(atomicEffect)
	case "SEARCH_CARD":
		return "カードをサーチ"
	case "SHUFFLE_DECK":
		return "デッキをシャッフル"
	case "MODIFY_ATTACK":
		return generateModifyStatDescription(atomicEffect)
	case "MODIFY_DEFENSE":
		return generateModifyStatDescription(atomicEffect)
	case "MODIFY_COST":
		return generateModifyCostDescription(atomicEffect)
	case "MODIFY_MAX_HP":
		return generateModifyMaxHPDescription(atomicEffect)
	case "SUMMON_UNIT":
		return generateSummonUnitDescription(atomicEffect)
	case "DESTROY_UNIT":
		return "ユニットを破壊"
	case "RETURN_TO_HAND":
		return "手札に戻す"
	case "RETURN_TO_DECK":
		return "デッキに戻す"
	case "DISABLE_UNIT":
		return "ユニットを無効化"
	case "GRANT_TRAIT":
		return generateGrantTraitDescription(atomicEffect)
	case "REMOVE_TRAIT":
		return generateRemoveTraitDescription(atomicEffect)
	case "GAIN_MANA":
		return generateGainManaDescription(atomicEffect)
	case "REDUCE_COST":
		return generateReduceCostDescription(atomicEffect)
	default:
		// 未知の効果タイプの場合は空文字列を返す
		return ""
	}
}

// generateDamageDescription ダメージ効果の説明を生成
func generateDamageDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "ダメージを与える"
	}

	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sに%dダメージ", targetDesc, amount)
	}
	return fmt.Sprintf("%dダメージ", amount)
}

// generateHealDescription 回復効果の説明を生成
func generateHealDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "回復する"
	}

	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sを%d回復", targetDesc, amount)
	}
	return fmt.Sprintf("%d回復", amount)
}

// generateDrawCardDescription カードドロー効果の説明を生成
func generateDrawCardDescription(atomicEffect *model.AtomicEffectModel) string {
	params := parseParameters(atomicEffect.Parameters)
	count, ok := params["count"].(float64)
	if !ok || count == 1 {
		return "カードを1枚引く"
	}
	return fmt.Sprintf("カードを%d枚引く", int(count))
}

// generateModifyStatDescription ステータス変更効果の説明を生成
func generateModifyStatDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	params := parseParameters(atomicEffect.Parameters)
	statType, _ := params["stat_type"].(string)

	if amount == 0 {
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "ステータスを変更する"
	}

	sign := ""
	if amount > 0 {
		sign = "+"
	}

	statName := "ステータス"
	switch atomicEffect.Type {
	case "MODIFY_ATTACK":
		statName = "攻撃力"
	case "MODIFY_DEFENSE":
		statName = "防御力"
	default:
		switch statType {
		case "ATTACK":
			statName = "攻撃力"
		case "DEFENSE":
			statName = "防御力"
		}
	}

	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sの%s%s%d", targetDesc, statName, sign, amount)
	}
	return fmt.Sprintf("%s%s%d", statName, sign, amount)
}

// generateDiscardDescription カード破棄効果の説明を生成
func generateDiscardDescription(atomicEffect *model.AtomicEffectModel) string {
	params := parseParameters(atomicEffect.Parameters)
	count, ok := params["count"].(float64)
	if !ok || count == 1 {
		return "カードを1枚捨てる"
	}
	return fmt.Sprintf("カードを%d枚捨てる", int(count))
}

// generateSplashDamageDescription 範囲ダメージ効果の説明を生成
func generateSplashDamageDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "範囲ダメージ"
	}

	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sに%d範囲ダメージ", targetDesc, amount)
	}
	return fmt.Sprintf("%d範囲ダメージ", amount)
}

// generateRestoreManaDescription マナ回復効果の説明を生成
func generateRestoreManaDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "マナを回復"
	}
	return fmt.Sprintf("マナを%d回復", amount)
}

// generateModifyCostDescription コスト変更効果の説明を生成
func generateModifyCostDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	sign := ""
	if amount > 0 {
		sign = "+"
	}

	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sのコスト%s%d", targetDesc, sign, amount)
	}
	return fmt.Sprintf("コスト%s%d", sign, amount)
}

// generateSummonUnitDescription ユニット召喚効果の説明を生成
func generateSummonUnitDescription(atomicEffect *model.AtomicEffectModel) string {
	params := parseParameters(atomicEffect.Parameters)

	// 召喚数を取得
	count := 1
	if countFloat, ok := params["count"].(float64); ok && countFloat > 0 {
		count = int(countFloat)
	}

	// ユニット名を取得
	unitName := ""
	if name, ok := params["name"].(string); ok && name != "" {
		unitName = name
	} else if cardID, ok := params["card_id"].(string); ok && cardID != "" {
		unitName = cardID
	}

	// 攻撃力/防御力を取得
	var attack, defense *int
	if attackFloat, ok := params["attack"].(float64); ok {
		val := int(attackFloat)
		attack = &val
	}
	if defenseFloat, ok := params["defense"].(float64); ok {
		val := int(defenseFloat)
		defense = &val
	}

	// 説明文を生成
	if unitName != "" {
		if attack != nil && defense != nil {
			return fmt.Sprintf("%s(%d/%d)を%d体召喚", unitName, *attack, *defense, count)
		}
		return fmt.Sprintf("%sを%d体召喚", unitName, count)
	}

	// ユニット名が指定されていない場合
	if attack != nil && defense != nil {
		return fmt.Sprintf("%d/%dのユニットを%d体召喚", *attack, *defense, count)
	}
	return fmt.Sprintf("ユニットを%d体召喚", count)
}

// generateModifyMaxHPDescription 最大HP変更効果の説明を生成
func generateModifyMaxHPDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	sign := ""
	if amount > 0 {
		sign = "+"
	}

	return fmt.Sprintf("最大HP%s%d", sign, amount)
}

// generateGrantTraitDescription 特性付与効果の説明を生成
func generateGrantTraitDescription(atomicEffect *model.AtomicEffectModel) string {
	params := parseParameters(atomicEffect.Parameters)
	trait, ok := params["trait"].(string)
	if !ok {
		return "特性を付与"
	}

	traitName := translateTraitName(trait)
	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sに%sを付与", targetDesc, traitName)
	}
	return fmt.Sprintf("%sを付与", traitName)
}

// generateRemoveTraitDescription 特性削除効果の説明を生成
func generateRemoveTraitDescription(atomicEffect *model.AtomicEffectModel) string {
	params := parseParameters(atomicEffect.Parameters)
	trait, ok := params["trait"].(string)
	if !ok {
		return "特性を削除"
	}

	traitName := translateTraitName(trait)
	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sから%sを削除", targetDesc, traitName)
	}
	return fmt.Sprintf("%sを削除", traitName)
}

// generateGainManaDescription マナ獲得効果の説明を生成
func generateGainManaDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "マナを獲得"
	}
	return fmt.Sprintf("マナを%d獲得", amount)
}

// generateReduceCostDescription コスト軽減効果の説明を生成
func generateReduceCostDescription(atomicEffect *model.AtomicEffectModel) string {
	amount := atomicEffect.Value
	if amount == 0 {
		params := parseParameters(atomicEffect.Parameters)
		if amountFloat, ok := params["amount"].(float64); ok {
			amount = int(amountFloat)
		}
	}

	if amount == 0 {
		return "コストを軽減"
	}

	targetDesc := generateTargetDescription(atomicEffect.Target)
	if targetDesc != "" {
		return fmt.Sprintf("%sのコストを%d軽減", targetDesc, amount)
	}
	return fmt.Sprintf("コストを%d軽減", amount)
}

// translateTraitName 特性名を日本語に変換
func translateTraitName(trait string) string {
	switch trait {
	case "RUSH":
		return "速攻"
	case "CHARGE":
		return "突撃"
	case "WINDFURY":
		return "疾風"
	case "PIERCE":
		return "貫通"
	case "GUARDIAN":
		return "守護"
	case "EFFECT_SHIELD":
		return "効果無効"
	case "UNTARGETABLE":
		return "対象不可"
	default:
		return trait
	}
}

// generateTargetDescription TargetSelectorから対象の説明を生成
func generateTargetDescription(selector *model.TargetSelectorModel) string {
	if selector == nil {
		return ""
	}

	switch selector.Type {
	case "Self":
		return "自身"
	case "Opponent":
		return "相手"
	case "Allies":
		return "味方全体"
	case "Enemies":
		return "敵全体"
	case "AllUnits":
		return "全てのユニット"
	case "RandomAlly":
		return "ランダムな味方1体"
	case "RandomEnemy":
		return "ランダムな敵1体"
	case "Specific":
		return "対象1体"
	default:
		return ""
	}
}

// generateAllTargetDescription 全体対象の説明を生成
func generateAllTargetDescription(selector *model.TargetSelectorModel) string {
	if selector.Filter == nil {
		switch selector.Type {
		case "Allies":
			return "味方全体"
		case "Enemies":
			return "敵全体"
		case "AllUnits":
			return "全てのユニット"
		default:
			return "全て"
		}
	}

	// フィルタがある場合の説明（簡略化）
	return "対象"
}

// generateRandomTargetDescription ランダム対象の説明を生成
func generateRandomTargetDescription(selector *model.TargetSelectorModel) string {
	count := selector.Count
	if count <= 1 {
		return "ランダムな対象1体"
	}
	return fmt.Sprintf("ランダムな対象%d体", count)
}

// generateSequenceDescription 順次実行の説明を生成（THEN）
func generateSequenceDescription(node *model.EffectChainNodeModel) string {
	// Sequentialノードの場合、Nextを辿る
	if node.Sequential == nil {
		return ""
	}

	var descriptions []string

	// 次のノードを処理
	if node.Sequential.Next != nil {
		desc := generateNodeDescription(node.Sequential.Next)
		if desc != "" {
			descriptions = append(descriptions, desc)
		}
	}

	return strings.Join(descriptions, "、")
}

// generateChoiceDescription 並列実行の説明を生成（AND）
func generateChoiceDescription(node *model.EffectChainNodeModel) string {
	// Parallelノードの場合、Childrenを取得
	if node.Parallel == nil || len(node.Parallel.Children) == 0 {
		return ""
	}

	var descriptions []string
	for _, child := range node.Parallel.Children {
		desc := generateNodeDescription(child)
		if desc != "" {
			descriptions = append(descriptions, desc)
		}
	}

	return strings.Join(descriptions, "、")
}

// generateIfElseDescription 条件分岐の説明を生成（IF_ELSE）
func generateIfElseDescription(node *model.EffectChainNodeModel) string {
	if node.IfElse == nil {
		return ""
	}

	// 簡略化: Thenブランチのみ表示
	if node.IfElse.Then != nil {
		return generateNodeDescription(node.IfElse.Then)
	}
	return ""
}

// generateRepeatDescription 繰り返しの説明を生成（REPEAT）
func generateRepeatDescription(node *model.EffectChainNodeModel) string {
	if node.Repeat == nil || node.Repeat.RepeatEffect == nil {
		return ""
	}

	desc := generateNodeDescription(node.Repeat.RepeatEffect)
	if desc == "" {
		return ""
	}

	count := node.Repeat.RepeatCount
	if count > 1 {
		return fmt.Sprintf("%s（%d回）", desc, count)
	}
	return desc
}

// generateForEachDescription 各対象への反復の説明を生成（FOREACH）
func generateForEachDescription(node *model.EffectChainNodeModel) string {
	if node.ForEach == nil || node.ForEach.ForEachEffect == nil {
		return ""
	}

	desc := generateNodeDescription(node.ForEach.ForEachEffect)
	if desc == "" {
		return ""
	}

	targetDesc := ""
	if node.ForEach.ForEachTarget != nil {
		targetDesc = generateTargetDescription(node.ForEach.ForEachTarget)
	}

	if targetDesc != "" {
		return fmt.Sprintf("%sに対して%s", targetDesc, desc)
	}
	return desc
} // parseParameters JSONパラメータをパース
func parseParameters(parametersJSON string) map[string]interface{} {
	if parametersJSON == "" {
		return make(map[string]interface{})
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(parametersJSON), &params); err != nil {
		return make(map[string]interface{})
	}
	return params
}

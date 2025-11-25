package entity

import "fmt"

// ========================================
// ゲーム内で使用する効果を定義する
// 設計方針: 原子効果を組み合わせて複雑な効果を構築する。原子効果同士は演算子で繋げることで複雑な効果を構築することを想定している。
// ========================================

// 原子効果種別
// 原始効果とは、カードの効果を表す最小単位の効果のこと。
type AtomicEffectType string

const (
	// ダメージ系
	AtomicEffectDealDamage AtomicEffectType = "DEAL_DAMAGE" // ダメージを与える
	AtomicEffectDealSplash AtomicEffectType = "DEAL_SPLASH" // 範囲ダメージ

	// 回復系
	AtomicEffectRestoreHP   AtomicEffectType = "RESTORE_HP"   // HP回復
	AtomicEffectRestoreMana AtomicEffectType = "RESTORE_MANA" // マナ回復
	AtomicEffectFullRestore AtomicEffectType = "FULL_RESTORE" // 完全回復

	// ドロー・デッキ操作系
	AtomicEffectDrawCard    AtomicEffectType = "DRAW_CARD"    // カードドロー
	AtomicEffectDiscardCard AtomicEffectType = "DISCARD_CARD" // 手札を捨てる
	AtomicEffectSearchCard  AtomicEffectType = "SEARCH_CARD"  // デッキからサーチ
	AtomicEffectShuffleDeck AtomicEffectType = "SHUFFLE_DECK" // デッキシャッフル

	// バフ・デバフ系
	AtomicEffectModifyAttack  AtomicEffectType = "MODIFY_ATTACK"  // 攻撃力変更
	AtomicEffectModifyDefense AtomicEffectType = "MODIFY_DEFENSE" // 防御力変更
	AtomicEffectModifyCost    AtomicEffectType = "MODIFY_COST"    // コスト変更
	AtomicEffectModifyMaxHP   AtomicEffectType = "MODIFY_MAX_HP"  // 最大HP変更

	// ユニット操作系
	AtomicEffectSummonUnit   AtomicEffectType = "SUMMON_UNIT"    // ユニット召喚
	AtomicEffectDestroyUnit  AtomicEffectType = "DESTROY_UNIT"   // ユニット破壊
	AtomicEffectReturnToHand AtomicEffectType = "RETURN_TO_HAND" // 手札に戻す
	AtomicEffectReturnToDeck AtomicEffectType = "RETURN_TO_DECK" // デッキに戻す
	AtomicEffectSilenceUnit  AtomicEffectType = "DISABLE_UNIT"   // 効果無効化

	// 特性操作系
	AtomicEffectGrantTrait  AtomicEffectType = "GRANT_TRAIT"  // 特性付与
	AtomicEffectRemoveTrait AtomicEffectType = "REMOVE_TRAIT" // 特性除去

	// リソース操作系
	AtomicEffectGainMana   AtomicEffectType = "GAIN_MANA"   // マナ増加
	AtomicEffectReduceCost AtomicEffectType = "REDUCE_COST" // コスト減少

)

// 効果演算子
// 効果演算子とは、効果をどう繋げるかを表す演算子のこと。
type EffectOperator string

const (
	OperatorSequential EffectOperator = "THEN"    // 順次実行: Aが完了したらB
	OperatorParallel   EffectOperator = "AND"     // 並列実行: AとBを独立して実行
	OperatorIfElse     EffectOperator = "IF_ELSE" // 条件分岐: IF 条件 THEN A ELSE B
	OperatorRepeat     EffectOperator = "REPEAT"  // 繰り返し: REPEAT(A, N回)
	OperatorForEach    EffectOperator = "FOREACH" // 反復: FOREACH 対象集合 DO A
	OperatorChoice     EffectOperator = "CHOOSE"  // 選択: プレイヤーが選択肢から選択
)

// 効果の対象
type EffectTarget string

const (
	EffectTargetSelf        EffectTarget = "Self"        // 自分自身
	EffectTargetOpponent    EffectTarget = "Opponent"    // 相手プレイヤー
	EffectTargetAllies      EffectTarget = "Allies"      // 自分側の全ユニット
	EffectTargetEnemies     EffectTarget = "Enemies"     // 相手側の全ユニット
	EffectTargetAllUnits    EffectTarget = "AllUnits"    // 両者の全ユニット
	EffectTargetRandomAlly  EffectTarget = "RandomAlly"  // ランダムな味方ユニット
	EffectTargetRandomEnemy EffectTarget = "RandomEnemy" // ランダムな敵ユニット
	EffectTargetSpecific    EffectTarget = "Specific"    // 明示的に指定されたユニット
)

// 効果の発動タイミング
type EffectTiming string

const (
	EffectTimingImmediate EffectTiming = "Immediate" // 即時発動
	EffectTimingOnSummon  EffectTiming = "OnSummon"  // 召喚時に発動
	EffectTimingOnDestroy EffectTiming = "OnDestroy" // 破壊時に発動
	EffectTimingOnAttack  EffectTiming = "OnAttack"  // 攻撃時に発動
	EffectTimingOnDamaged EffectTiming = "OnDamaged" // ダメージを受けたときに発動
	EffectTimingTurnStart EffectTiming = "TurnStart" // ターン開始時
	EffectTimingTurnEnd   EffectTiming = "TurnEnd"   // ターン終了時
)

// 対象選択
type TargetSelector struct {
	Type        EffectTarget  `json:"type"`                    // 基本対象タイプ
	Filter      *TargetFilter `json:"filter,omitempty"`        // フィルタ条件
	Count       int           `json:"count,omitempty"`         // 選択数（0=全て）
	Random      bool          `json:"random,omitempty"`        // ランダム選択
	SelectByMax string        `json:"select_by_max,omitempty"` // 最大値で選択する属性（"attack", "defense", "cost"）
	SelectByMin string        `json:"select_by_min,omitempty"` // 最小値で選択する属性（"attack", "defense", "cost"）
}

// 対象フィルタ
type TargetFilter struct {
	MinAttack  *int      `json:"min_attack,omitempty"`   // 最小攻撃力
	MaxAttack  *int      `json:"max_attack,omitempty"`   // 最大攻撃力
	MinDefense *int      `json:"min_defense,omitempty"`  // 最小防御力
	MaxDefense *int      `json:"max_defense,omitempty"`  // 最大防御力
	MinCost    *int      `json:"min_cost,omitempty"`     // 最小コスト
	MaxCost    *int      `json:"max_cost,omitempty"`     // 最大コスト
	HasTrait   []Trait   `json:"has_keyword,omitempty"`  // 持っている特性
	LackTrait  []Trait   `json:"lack_keyword,omitempty"` // 持っていない特性
	CardType   *CardType `json:"card_type,omitempty"`    // カードタイプ
}

// 条件定義
type Condition struct {
	Type     ConditionType      `json:"type"`               // 条件タイプ
	Operator ComparisonOperator `json:"operator,omitempty"` // 比較演算子
	Value    int                `json:"value,omitempty"`    // 比較値
}

// 比較演算子
type ComparisonOperator string

const (
	OperatorEqual              ComparisonOperator = "EQUAL"                 // ==
	OperatorNotEqual           ComparisonOperator = "NOT_EQUAL"             // !=
	OperatorLessThan           ComparisonOperator = "LESS_THAN"             // <
	OperatorGreaterThan        ComparisonOperator = "GREATER_THAN"          // >
	OperatorLessThanOrEqual    ComparisonOperator = "LESS_THAN_OR_EQUAL"    // <=
	OperatorGreaterThanOrEqual ComparisonOperator = "GREATER_THAN_OR_EQUAL" // >=
)

// 条件タイプ
type ConditionType string

const (
	ConditionPlayerHP    ConditionType = "PLAYER_HP"    // プレイヤーHP
	ConditionPlayerMana  ConditionType = "PLAYER_MANA"  // プレイヤーマナ
	ConditionUnitCount   ConditionType = "UNIT_COUNT"   // ユニット数
	ConditionHandSize    ConditionType = "HAND_SIZE"    // 手札枚数
	ConditionDeckSize    ConditionType = "DECK_SIZE"    // デッキ枚数
	ConditionTurnNumber  ConditionType = "TURN_NUMBER"  // ターン数
	ConditionUnitAttack  ConditionType = "UNIT_ATTACK"  // ユニット攻撃力
	ConditionUnitDefense ConditionType = "UNIT_DEFENSE" // ユニット防御力
	ConditionHasTrait    ConditionType = "HAS_KEYWORD"  // 特性所持
	ConditionCardPlayed  ConditionType = "CARD_PLAYED"  // このターンプレイしたカード数
	ConditionDamageTaken ConditionType = "DAMAGE_TAKEN" // 受けたダメージ量
)

// 原子効果
// Timing: 発動タイミング（召喚時、攻撃時など、条件が満たされているかどうか）
// Condition: 発動条件（HPが10以下など、状態が満たされているかどうか）
// 注意: TimingがImmediate以外の場合、Conditionは無視される
// 実装: ProcessTimingEffectsでTimingをチェックし、executeAtomicEffectでConditionをチェック
type AtomicEffect struct {
	ID         string           `json:"id"`                   // 効果ID（DB用）
	Type       AtomicEffectType `json:"type"`                 // 効果タイプ
	Target     TargetSelector   `json:"target"`               // 対象選択
	Value      int              `json:"value,omitempty"`      // 効果値
	Multiplier float64          `json:"multiplier,omitempty"` // 倍率（デフォルト1.0）
	Duration   *int             `json:"duration,omitempty"`   // 持続ターン数
	Timing     EffectTiming     `json:"timing"`               // 発動タイミング
	Condition  *Condition       `json:"condition,omitempty"`  // 発動条件
	Parameters map[string]any   `json:"parameters,omitempty"` // 追加パラメータ
}

// 演算子ごとのノード型（型安全性のため）

// SequentialNode 順次実行ノード: Aが完了したらBを実行
type SequentialNode struct {
	Effect *AtomicEffect    `json:"effect,omitempty"` // 原子効果（省略可能）
	Next   *EffectChainNode `json:"next,omitempty"`   // 次のノード（省略可能、nilの場合は終了）
}

// ParallelNode 並列実行ノード: AとBを独立して実行
type ParallelNode struct {
	Children []*EffectChainNode `json:"children"`       // 子ノード（必須）
	Next     *EffectChainNode   `json:"next,omitempty"` // 並列実行後の次のノード（省略可能）
}

// IfElseNode 条件分岐ノード: IF 条件 THEN A ELSE B
type IfElseNode struct {
	Condition *Condition       `json:"condition"`      // 条件（必須）
	Then      *EffectChainNode `json:"then"`           // 条件が真の場合のノード（必須）
	Else      *EffectChainNode `json:"else,omitempty"` // 条件が偽の場合のノード（省略可能）
}

// RepeatNode 繰り返しノード: REPEAT(A, N回)
type RepeatNode struct {
	Count  int              `json:"count"`  // 繰り返し回数（必須）
	Effect *EffectChainNode `json:"effect"` // 繰り返す効果（必須）
}

// ForEachNode 反復ノード: FOREACH 対象集合 DO A
type ForEachNode struct {
	Target TargetSelector   `json:"target"` // 対象選択（必須）
	Effect *EffectChainNode `json:"effect"` // 各対象に適用する効果（必須）
}

// ChoiceNode 選択ノード: プレイヤーが選択肢から選択（未実装）
type ChoiceNode struct {
	Options []*EffectChainNode `json:"options"` // 選択肢（必須）
}

// 効果チェーンのノード
// 型安全性を保つため、演算子ごとに専用の構造体を使用
type EffectChainNode struct {
	Type EffectOperator `json:"type"` // 演算子タイプ（必須）

	// 演算子ごとのデータ（型安全にアクセスするためのヘルパーメソッドを提供）
	Sequential *SequentialNode `json:"sequential,omitempty"`
	Parallel   *ParallelNode   `json:"parallel,omitempty"`
	IfElse     *IfElseNode     `json:"if_else,omitempty"`
	Repeat     *RepeatNode     `json:"repeat,omitempty"`
	ForEach    *ForEachNode    `json:"for_each,omitempty"`
	Choice     *ChoiceNode     `json:"choice,omitempty"`
}

// 型安全なアクセサメソッド

// GetSequential SequentialNodeを取得（型安全）
func (n *EffectChainNode) GetSequential() (*SequentialNode, bool) {
	return n.Sequential, n.Type == OperatorSequential && n.Sequential != nil
}

// GetParallel ParallelNodeを取得（型安全）
func (n *EffectChainNode) GetParallel() (*ParallelNode, bool) {
	return n.Parallel, n.Type == OperatorParallel && n.Parallel != nil
}

// GetIfElse IfElseNodeを取得（型安全）
func (n *EffectChainNode) GetIfElse() (*IfElseNode, bool) {
	return n.IfElse, n.Type == OperatorIfElse && n.IfElse != nil
}

// GetRepeat RepeatNodeを取得（型安全）
func (n *EffectChainNode) GetRepeat() (*RepeatNode, bool) {
	return n.Repeat, n.Type == OperatorRepeat && n.Repeat != nil
}

// GetForEach ForEachNodeを取得（型安全）
func (n *EffectChainNode) GetForEach() (*ForEachNode, bool) {
	return n.ForEach, n.Type == OperatorForEach && n.ForEach != nil
}

// GetChoice ChoiceNodeを取得（型安全）
func (n *EffectChainNode) GetChoice() (*ChoiceNode, bool) {
	return n.Choice, n.Type == OperatorChoice && n.Choice != nil
}

// IsValid ノードが有効かどうかを判定（Typeと対応するフィールドが一致しているか）
// validatorで使用することを想定
func (n *EffectChainNode) IsValid() bool {
	if n == nil {
		return false
	}

	switch n.Type {
	case OperatorSequential:
		return n.Sequential != nil
	case OperatorParallel:
		return n.Parallel != nil
	case OperatorIfElse:
		return n.IfElse != nil
	case OperatorRepeat:
		return n.Repeat != nil
	case OperatorForEach:
		return n.ForEach != nil
	case OperatorChoice:
		return n.Choice != nil
	default:
		return false
	}
}

// HasMultipleOperators 複数の演算子フィールドが設定されているか（無効な状態）
// validatorで使用することを想定
func (n *EffectChainNode) HasMultipleOperators() bool {
	if n == nil {
		return false
	}

	count := 0
	if n.Sequential != nil {
		count++
	}
	if n.Parallel != nil {
		count++
	}
	if n.IfElse != nil {
		count++
	}
	if n.Repeat != nil {
		count++
	}
	if n.ForEach != nil {
		count++
	}
	if n.Choice != nil {
		count++
	}

	return count > 1
}

// 効果定義
// 1つのカード効果を定義する構造体（原子効果と演算子を組み合わせて複雑な効果を表現）
type EffectDefinition struct {
	ID            string           `json:"id"`             // 定義ID（DB用）
	Name          string           `json:"name"`           // 効果名
	Root          *EffectChainNode `json:"root"`           // ルートノード
	RequireTarget bool             `json:"require_target"` // 対象選択必須
}

// カードが持つ効果の定義
type CardEffect struct {
	Definitions []*EffectDefinition `json:"definitions"` // 効果定義のリスト
	Description string              `json:"description"` // 統合された説明文
}

// ========================================
// 効果説明文の自動生成
// ========================================

// GenerateDescription 効果定義から説明文を自動生成
func (e *EffectDefinition) GenerateDescription() string {
	if e.Root == nil {
		return ""
	}
	return generateNodeDescription(e.Root)
}

// GenerateDescription CardEffectから統合された説明文を自動生成
func (c *CardEffect) GenerateDescription() string {
	if len(c.Definitions) == 0 {
		return ""
	}

	descriptions := make([]string, 0, len(c.Definitions))
	for _, def := range c.Definitions {
		if desc := def.GenerateDescription(); desc != "" {
			descriptions = append(descriptions, desc)
		}
	}

	if len(descriptions) == 0 {
		return ""
	}
	if len(descriptions) == 1 {
		return descriptions[0]
	}

	// 複数の効果定義がある場合は改行で連結
	result := ""
	for i, desc := range descriptions {
		if i > 0 {
			result += "\n"
		}
		result += desc
	}
	return result
}

// generateNodeDescription ノードから説明文を生成
func generateNodeDescription(node *EffectChainNode) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case OperatorSequential:
		return generateSequentialDescription(node.Sequential)
	case OperatorParallel:
		return generateParallelDescription(node.Parallel)
	case OperatorIfElse:
		return generateIfElseDescription(node.IfElse)
	case OperatorRepeat:
		return generateRepeatDescription(node.Repeat)
	case OperatorForEach:
		return generateForEachDescription(node.ForEach)
	case OperatorChoice:
		return generateChoiceDescription(node.Choice)
	default:
		return ""
	}
}

// generateSequentialDescription 順次実行ノードの説明文を生成
func generateSequentialDescription(node *SequentialNode) string {
	if node == nil {
		return ""
	}

	parts := []string{}

	// 現在の効果を追加
	if node.Effect != nil {
		if desc := generateAtomicEffectDescription(node.Effect); desc != "" {
			parts = append(parts, desc)
		}
	}

	// 次のノードを追加
	if node.Next != nil {
		if desc := generateNodeDescription(node.Next); desc != "" {
			parts = append(parts, desc)
		}
	}

	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	// 順次実行は「、」で連結
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "、"
		}
		result += part
	}
	return result
}

// generateParallelDescription 並列実行ノードの説明文を生成
func generateParallelDescription(node *ParallelNode) string {
	if node == nil || len(node.Children) == 0 {
		return ""
	}

	parts := []string{}
	for _, child := range node.Children {
		if desc := generateNodeDescription(child); desc != "" {
			parts = append(parts, desc)
		}
	}

	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	// 並列実行は「、」で連結
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "、"
		}
		result += part
	}
	return result
}

// generateIfElseDescription 条件分岐ノードの説明文を生成
func generateIfElseDescription(node *IfElseNode) string {
	if node == nil {
		return ""
	}

	condDesc := generateConditionDescription(node.Condition)
	thenDesc := generateNodeDescription(node.Then)

	if node.Else == nil {
		return condDesc + "なら" + thenDesc
	}

	elseDesc := generateNodeDescription(node.Else)
	return condDesc + "なら" + thenDesc + "、そうでなければ" + elseDesc
}

// generateRepeatDescription 繰り返しノードの説明文を生成
func generateRepeatDescription(node *RepeatNode) string {
	if node == nil {
		return ""
	}

	effectDesc := generateNodeDescription(node.Effect)
	return fmt.Sprintf("%sを%d回行う", effectDesc, node.Count)
}

// generateForEachDescription 反復ノードの説明文を生成
func generateForEachDescription(node *ForEachNode) string {
	if node == nil {
		return ""
	}

	targetDesc := generateTargetDescription(node.Target.Type)
	effectDesc := generateNodeDescription(node.Effect)
	return fmt.Sprintf("%s1体につき、%s", targetDesc, effectDesc)
}

// generateChoiceDescription 選択ノードの説明文を生成
func generateChoiceDescription(node *ChoiceNode) string {
	if node == nil || len(node.Options) == 0 {
		return ""
	}

	parts := []string{}
	for _, option := range node.Options {
		if desc := generateNodeDescription(option); desc != "" {
			parts = append(parts, desc)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	result := "次のいずれかを選択: "
	for i, part := range parts {
		if i > 0 {
			result += "、または"
		}
		result += part
	}
	return result
}

// generateAtomicEffectDescription 原子効果の説明文を生成
func generateAtomicEffectDescription(effect *AtomicEffect) string {
	if effect == nil {
		return ""
	}

	targetDesc := generateTargetDescription(effect.Target.Type)
	valueStr := ""
	if effect.Value > 0 {
		valueStr = fmt.Sprintf("%d", effect.Value)
	}

	switch effect.Type {
	case AtomicEffectDealDamage:
		return fmt.Sprintf("%sに%sダメージ", targetDesc, valueStr)
	case AtomicEffectDealSplash:
		return fmt.Sprintf("%sに%sダメージ", targetDesc, valueStr)
	case AtomicEffectRestoreHP:
		return fmt.Sprintf("%sのHPを%s回復", targetDesc, valueStr)
	case AtomicEffectRestoreMana:
		return fmt.Sprintf("マナを%s回復", valueStr)
	case AtomicEffectFullRestore:
		return fmt.Sprintf("%sを完全回復", targetDesc)
	case AtomicEffectDrawCard:
		return fmt.Sprintf("カードを%s枚引く", valueStr)
	case AtomicEffectDiscardCard:
		return fmt.Sprintf("カードを%s枚捨てる", valueStr)
	case AtomicEffectSearchCard:
		return "デッキからカードをサーチ"
	case AtomicEffectShuffleDeck:
		return "デッキをシャッフル"
	case AtomicEffectModifyAttack:
		sign := "+"
		if effect.Value < 0 {
			sign = ""
		}
		return fmt.Sprintf("%sの攻撃力%s%d", targetDesc, sign, effect.Value)
	case AtomicEffectModifyDefense:
		sign := "+"
		if effect.Value < 0 {
			sign = ""
		}
		return fmt.Sprintf("%sの防御力%s%d", targetDesc, sign, effect.Value)
	case AtomicEffectModifyCost:
		sign := "+"
		if effect.Value < 0 {
			sign = ""
		}
		return fmt.Sprintf("%sのコスト%s%d", targetDesc, sign, effect.Value)
	case AtomicEffectModifyMaxHP:
		sign := "+"
		if effect.Value < 0 {
			sign = ""
		}
		return fmt.Sprintf("%sの最大HP%s%d", targetDesc, sign, effect.Value)
	case AtomicEffectSummonUnit:
		return "ユニットを召喚"
	case AtomicEffectDestroyUnit:
		return fmt.Sprintf("%sを破壊", targetDesc)
	case AtomicEffectReturnToHand:
		return fmt.Sprintf("%sを手札に戻す", targetDesc)
	case AtomicEffectReturnToDeck:
		return fmt.Sprintf("%sをデッキに戻す", targetDesc)
	case AtomicEffectSilenceUnit:
		return fmt.Sprintf("%sの効果を無効化", targetDesc)
	case AtomicEffectGrantTrait:
		if effect.Parameters != nil {
			if trait, ok := effect.Parameters["trait"].(Trait); ok {
				traitDesc := generateTraitDescription(trait)
				return fmt.Sprintf("%sに%sを付与", targetDesc, traitDesc)
			}
		}
		return fmt.Sprintf("%sに特性を付与", targetDesc)
	case AtomicEffectRemoveTrait:
		if effect.Parameters != nil {
			if trait, ok := effect.Parameters["trait"].(Trait); ok {
				traitDesc := generateTraitDescription(trait)
				return fmt.Sprintf("%sから%sを除去", targetDesc, traitDesc)
			}
		}
		return fmt.Sprintf("%sから特性を除去", targetDesc)
	case AtomicEffectGainMana:
		return fmt.Sprintf("マナ+%d", effect.Value)
	case AtomicEffectReduceCost:
		return fmt.Sprintf("コスト-%d", effect.Value)
	default:
		return "未定義の効果"
	}
}

// generateTargetDescription 対象タイプの説明文を生成
func generateTargetDescription(target EffectTarget) string {
	switch target {
	case EffectTargetSelf:
		return "自分"
	case EffectTargetOpponent:
		return "相手"
	case EffectTargetAllies:
		return "味方全体"
	case EffectTargetEnemies:
		return "敵全体"
	case EffectTargetAllUnits:
		return "全ユニット"
	case EffectTargetRandomAlly:
		return "ランダムな味方"
	case EffectTargetRandomEnemy:
		return "ランダムな敵"
	case EffectTargetSpecific:
		return "対象"
	default:
		return "対象"
	}
}

// generateConditionDescription 条件の説明文を生成
func generateConditionDescription(cond *Condition) string {
	if cond == nil {
		return ""
	}

	var targetStr string
	switch cond.Type {
	case ConditionPlayerHP:
		targetStr = "自分のHP"
	case ConditionPlayerMana:
		targetStr = "自分のマナ"
	case ConditionUnitCount:
		targetStr = "ユニット数"
	case ConditionHandSize:
		targetStr = "手札枚数"
	case ConditionDeckSize:
		targetStr = "デッキ枚数"
	case ConditionTurnNumber:
		targetStr = "ターン数"
	case ConditionUnitAttack:
		targetStr = "ユニット攻撃力"
	case ConditionUnitDefense:
		targetStr = "ユニット防御力"
	case ConditionHasTrait:
		targetStr = "特性を持つ"
	case ConditionCardPlayed:
		targetStr = "プレイしたカード数"
	case ConditionDamageTaken:
		targetStr = "受けたダメージ"
	default:
		targetStr = "条件"
	}

	var opStr string
	switch cond.Operator {
	case OperatorEqual:
		opStr = "が"
	case OperatorNotEqual:
		opStr = "が"
	case OperatorLessThan:
		opStr = "が"
	case OperatorGreaterThan:
		opStr = "が"
	case OperatorLessThanOrEqual:
		opStr = "が"
	case OperatorGreaterThanOrEqual:
		opStr = "が"
	default:
		opStr = "が"
	}

	valueStr := fmt.Sprintf("%d", cond.Value)

	switch cond.Operator {
	case OperatorEqual:
		return fmt.Sprintf("%s%s%s", targetStr, opStr, valueStr)
	case OperatorNotEqual:
		return fmt.Sprintf("%s%s%sでない", targetStr, opStr, valueStr)
	case OperatorLessThan:
		return fmt.Sprintf("%s%s%s未満", targetStr, opStr, valueStr)
	case OperatorGreaterThan:
		return fmt.Sprintf("%s%s%s超", targetStr, opStr, valueStr)
	case OperatorLessThanOrEqual:
		return fmt.Sprintf("%s%s%s以下", targetStr, opStr, valueStr)
	case OperatorGreaterThanOrEqual:
		return fmt.Sprintf("%s%s%s以上", targetStr, opStr, valueStr)
	default:
		return targetStr
	}
}

// generateTraitDescription 特性の説明文を生成
func generateTraitDescription(trait Trait) string {
	switch trait {
	case TraitRush:
		return "疾走"
	case TraitCharge:
		return "突進"
	case TraitWindfury:
		return "疾風"
	case TraitPierce:
		return "貫通"
	case TraitGuardian:
		return "守護"
	case TraitEffectShield:
		return "効果盾"
	case TraitUntargetable:
		return "対象不可"
	default:
		return string(trait)
	}
}

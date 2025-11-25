package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CardModel GORMモデル（データベース用）
type CardModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)"`
	Name      string    `gorm:"type:varchar(255);not null;index"`
	Type      string    `gorm:"type:enum('Unit','Spell','Leader');not null;index"`
	Cost      int       `gorm:"not null;index"`
	Attack    *int      `gorm:"type:int"`
	Defense   *int      `gorm:"type:int"`
	Effect    string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// リレーション
	Traits     []CardTraitModel `gorm:"foreignKey:CardID;constraint:OnDelete:CASCADE"`
	CardEffect *CardEffectModel `gorm:"foreignKey:CardID;constraint:OnDelete:CASCADE"`
}

// CardTraitModel カードと特性の中間テーブル
type CardTraitModel struct {
	CardID    string    `gorm:"primaryKey;type:varchar(255)"`
	Card      CardModel `gorm:"foreignKey:CardID;constraint:OnDelete:CASCADE"`
	Trait     string    `gorm:"type:enum('RUSH','CHARGE','WINDFURY','PIERCE','GUARDIAN','EFFECT_SHIELD','UNTARGETABLE');primaryKey;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// CardEffectModel カード効果のGORMモデル
type CardEffectModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	CardID    string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"card_id,omitempty"`
	Card      CardModel `gorm:"foreignKey:CardID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`

	// リレーション
	Definitions []EffectDefinitionModel `gorm:"foreignKey:CardEffectID;constraint:OnDelete:CASCADE" json:"definitions,omitempty"`
}

// EffectDefinitionModel 効果定義のGORMモデル
type EffectDefinitionModel struct {
	ID            uint                  `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	CardEffectID  uint                  `gorm:"not null;index" json:"card_effect_id,omitempty"`
	CardEffect    CardEffectModel       `gorm:"foreignKey:CardEffectID;constraint:OnDelete:CASCADE" json:"-"`
	RequireTarget bool                  `gorm:"default:false" json:"require_target,omitempty"`
	RootNodeID    *uint                 `gorm:"index" json:"root_node_id,omitempty"`
	Root          *EffectChainNodeModel `gorm:"foreignKey:RootNodeID;constraint:OnDelete:SET NULL" json:"root,omitempty"`
	CreatedAt     time.Time             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt     time.Time             `gorm:"autoUpdateTime" json:"-"`
}

// EffectChainNodeModel 効果チェーンノードのベーステーブル（共通フィールド）
type EffectChainNodeModel struct {
	ID             uint               `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Type           string             `gorm:"type:enum('THEN','AND','IF_ELSE','REPEAT','FOREACH');not null" json:"type,omitempty"`
	AtomicEffectID *uint              `gorm:"index" json:"atomic_effect_id,omitempty"`
	AtomicEffect   *AtomicEffectModel `gorm:"foreignKey:AtomicEffectID;constraint:OnDelete:SET NULL" json:"atomic_effect,omitempty"`
	CreatedAt      time.Time          `gorm:"autoCreateTime" json:"-"`
	UpdatedAt      time.Time          `gorm:"autoUpdateTime" json:"-"`

	// ノードタイプに応じた具体テーブルへの参照（JSONでは使用しない）
	Sequential *SequentialNodeModel `gorm:"-" json:"sequential,omitempty"`
	Parallel   *ParallelNodeModel   `gorm:"-" json:"parallel,omitempty"`
	IfElse     *IfElseNodeModel     `gorm:"-" json:"if_else,omitempty"`
	Repeat     *RepeatNodeModel     `gorm:"-" json:"repeat,omitempty"`
	ForEach    *ForEachNodeModel    `gorm:"-" json:"for_each,omitempty"`
}

// SequentialNodeModel 順次実行ノード（Type='THEN'の場合）
type SequentialNodeModel struct {
	NodeID    uint                  `gorm:"primaryKey" json:"node_id,omitempty"`
	Node      EffectChainNodeModel  `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE" json:"-"`
	NextID    *uint                 `gorm:"index" json:"next_id,omitempty"`
	Next      *EffectChainNodeModel `gorm:"foreignKey:NextID;constraint:OnDelete:SET NULL" json:"next,omitempty"`
	CreatedAt time.Time             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time             `gorm:"autoUpdateTime" json:"-"`
}

// ParallelNodeModel 並列実行ノード（Type='AND'の場合）
type ParallelNodeModel struct {
	NodeID         uint                  `gorm:"primaryKey" json:"node_id,omitempty"`
	Node           EffectChainNodeModel  `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE" json:"-"`
	ParallelNextID *uint                 `gorm:"index" json:"parallel_next_id,omitempty"`
	ParallelNext   *EffectChainNodeModel `gorm:"foreignKey:ParallelNextID;constraint:OnDelete:SET NULL" json:"parallel_next,omitempty"`
	CreatedAt      time.Time             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt      time.Time             `gorm:"autoUpdateTime" json:"-"`

	// Childrenは中間テーブルで管理
	Children []*EffectChainNodeModel `gorm:"many2many:parallel_node_children;" json:"children,omitempty"`
}

// ParallelNodeChildModel 並列ノードの子ノード中間テーブル
type ParallelNodeChildModel struct {
	ParallelNodeID uint                 `gorm:"primaryKey"`
	ParallelNode   ParallelNodeModel    `gorm:"foreignKey:ParallelNodeID;constraint:OnDelete:CASCADE"`
	ChildNodeID    uint                 `gorm:"primaryKey"`
	ChildNode      EffectChainNodeModel `gorm:"foreignKey:ChildNodeID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time            `gorm:"autoCreateTime"`
	UpdatedAt      time.Time            `gorm:"autoUpdateTime"`
}

// IfElseNodeModel 条件分岐ノード（Type='IF_ELSE'の場合）
type IfElseNodeModel struct {
	NodeID      uint                  `gorm:"primaryKey" json:"node_id,omitempty"`
	Node        EffectChainNodeModel  `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE" json:"-"`
	ThenID      uint                  `gorm:"not null;index" json:"then_id,omitempty"`
	Then        *EffectChainNodeModel `gorm:"foreignKey:ThenID;constraint:OnDelete:CASCADE" json:"then,omitempty"`
	ElseID      *uint                 `gorm:"index" json:"else_id,omitempty"`
	Else        *EffectChainNodeModel `gorm:"foreignKey:ElseID;constraint:OnDelete:SET NULL" json:"else,omitempty"`
	ConditionID uint                  `gorm:"not null;index" json:"condition_id,omitempty"`
	Condition   *ConditionModel       `gorm:"foreignKey:ConditionID;constraint:OnDelete:CASCADE" json:"condition,omitempty"`
	CreatedAt   time.Time             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt   time.Time             `gorm:"autoUpdateTime" json:"-"`
}

// RepeatNodeModel 繰り返しノード（Type='REPEAT'の場合）
type RepeatNodeModel struct {
	NodeID         uint                  `gorm:"primaryKey" json:"node_id,omitempty"`
	Node           EffectChainNodeModel  `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE" json:"-"`
	RepeatEffectID uint                  `gorm:"not null;index" json:"repeat_effect_id,omitempty"`
	RepeatEffect   *EffectChainNodeModel `gorm:"foreignKey:RepeatEffectID;constraint:OnDelete:CASCADE" json:"repeat_effect,omitempty"`
	RepeatCount    int                   `gorm:"not null" json:"count,omitempty"`
	CreatedAt      time.Time             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt      time.Time             `gorm:"autoUpdateTime" json:"-"`
}

// ForEachNodeModel 反復ノード（Type='FOREACH'の場合）
type ForEachNodeModel struct {
	NodeID          uint                  `gorm:"primaryKey" json:"node_id,omitempty"`
	Node            EffectChainNodeModel  `gorm:"foreignKey:NodeID;constraint:OnDelete:CASCADE" json:"-"`
	ForEachEffectID uint                  `gorm:"not null;index" json:"for_each_effect_id,omitempty"`
	ForEachEffect   *EffectChainNodeModel `gorm:"foreignKey:ForEachEffectID;constraint:OnDelete:CASCADE" json:"for_each_effect,omitempty"`
	ForEachTargetID uint                  `gorm:"not null;index" json:"for_each_target_id,omitempty"`
	ForEachTarget   *TargetSelectorModel  `gorm:"foreignKey:ForEachTargetID;constraint:OnDelete:CASCADE" json:"for_each_target,omitempty"`
	CreatedAt       time.Time             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt       time.Time             `gorm:"autoUpdateTime" json:"-"`
}

// AtomicEffectModel 原子効果のGORMモデル
type AtomicEffectModel struct {
	ID          uint                 `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Type        string               `gorm:"type:enum('DEAL_DAMAGE','DEAL_SPLASH','RESTORE_HP','RESTORE_MANA','FULL_RESTORE','DRAW_CARD','DISCARD_CARD','SEARCH_CARD','SHUFFLE_DECK','MODIFY_ATTACK','MODIFY_DEFENSE','MODIFY_COST','MODIFY_MAX_HP','SUMMON_UNIT','DESTROY_UNIT','RETURN_TO_HAND','RETURN_TO_DECK','DISABLE_UNIT','GRANT_TRAIT','REMOVE_TRAIT','GAIN_MANA','REDUCE_COST');not null" json:"type,omitempty"`
	Value       int                  `gorm:"default:0" json:"value,omitempty"`
	Multiplier  float64              `gorm:"type:double;default:1.0" json:"multiplier,omitempty"`
	Duration    *int                 `gorm:"type:int" json:"duration,omitempty"`
	Timing      string               `gorm:"type:enum('Immediate','OnSummon','OnDestroy','OnAttack','OnDamaged','TurnStart','TurnEnd');not null" json:"timing,omitempty"`
	TargetID    uint                 `gorm:"not null;index" json:"target_id,omitempty"`
	Target      *TargetSelectorModel `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"target,omitempty"`
	ConditionID *uint                `gorm:"index" json:"condition_id,omitempty"`
	Condition   *ConditionModel      `gorm:"foreignKey:ConditionID;constraint:OnDelete:SET NULL" json:"condition,omitempty"`
	Parameters  string               `gorm:"type:json" json:"parameters,omitempty"` // map[string]anyはJSONとして保存
	CreatedAt   time.Time            `gorm:"autoCreateTime" json:"-"`
	UpdatedAt   time.Time            `gorm:"autoUpdateTime" json:"-"`
}

// TargetSelectorModel 対象選択のGORMモデル
type TargetSelectorModel struct {
	ID          uint               `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Type        string             `gorm:"type:enum('Self','Opponent','Allies','Enemies','AllUnits','RandomAlly','RandomEnemy','Specific');not null" json:"type,omitempty"`
	Count       int                `gorm:"default:0" json:"count,omitempty"`
	Random      bool               `gorm:"default:false" json:"random,omitempty"`
	SelectByMax string             `gorm:"type:varchar(50)" json:"select_by_max,omitempty"`
	SelectByMin string             `gorm:"type:varchar(50)" json:"select_by_min,omitempty"`
	FilterID    *uint              `gorm:"index" json:"filter_id,omitempty"`
	Filter      *TargetFilterModel `gorm:"foreignKey:FilterID;constraint:OnDelete:SET NULL" json:"filter,omitempty"`
	CreatedAt   time.Time          `gorm:"autoCreateTime" json:"-"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime" json:"-"`
}

// TargetFilterModel 対象フィルタのGORMモデル
type TargetFilterModel struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	MinAttack  *int      `gorm:"type:int" json:"min_attack,omitempty"`
	MaxAttack  *int      `gorm:"type:int" json:"max_attack,omitempty"`
	MinDefense *int      `gorm:"type:int" json:"min_defense,omitempty"`
	MaxDefense *int      `gorm:"type:int" json:"max_defense,omitempty"`
	MinCost    *int      `gorm:"type:int" json:"min_cost,omitempty"`
	MaxCost    *int      `gorm:"type:int" json:"max_cost,omitempty"`
	CardType   string    `gorm:"type:enum('Unit','Spell','Leader')" json:"card_type,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"-"`

	// 特性は別テーブルで管理
	HasTraits  []TargetFilterTraitModel `gorm:"foreignKey:FilterID;constraint:OnDelete:CASCADE" json:"has_traits,omitempty"`
	LackTraits []TargetFilterTraitModel `gorm:"foreignKey:FilterID;constraint:OnDelete:CASCADE" json:"lack_traits,omitempty"`
}

// TargetFilterTraitModel フィルタ特性の中間テーブル
type TargetFilterTraitModel struct {
	FilterID   uint              `gorm:"primaryKey" json:"filter_id,omitempty"`
	Filter     TargetFilterModel `gorm:"foreignKey:FilterID;constraint:OnDelete:CASCADE" json:"-"`
	Trait      string            `gorm:"type:enum('RUSH','CHARGE','WINDFURY','PIERCE','GUARDIAN','EFFECT_SHIELD','UNTARGETABLE');primaryKey" json:"trait,omitempty"`
	IsHasTrait bool              `gorm:"not null" json:"is_has_trait,omitempty"` // true=HasTrait, false=LackTrait
	CreatedAt  time.Time         `gorm:"autoCreateTime" json:"-"`
	UpdatedAt  time.Time         `gorm:"autoUpdateTime" json:"-"`
}

// ConditionModel 条件のGORMモデル
type ConditionModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Type      string    `gorm:"type:enum('PLAYER_HP','PLAYER_MANA','UNIT_COUNT','HAND_SIZE','DECK_SIZE','TURN_NUMBER','UNIT_ATTACK','UNIT_DEFENSE','HAS_KEYWORD','CARD_PLAYED','DAMAGE_TAKEN');not null" json:"type,omitempty"`
	Operator  string    `gorm:"type:enum('EQUAL','NOT_EQUAL','LESS_THAN','GREATER_THAN','LESS_THAN_OR_EQUAL','GREATER_THAN_OR_EQUAL')" json:"operator,omitempty"`
	Value     int       `gorm:"default:0" json:"value,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`
}

// TableName テーブル名を指定
func (CardModel) TableName() string {
	return "cards"
}

func (CardTraitModel) TableName() string {
	return "card_traits"
}

func (CardEffectModel) TableName() string {
	return "card_effects"
}

func (EffectDefinitionModel) TableName() string {
	return "effect_definitions"
}

func (EffectChainNodeModel) TableName() string {
	return "effect_chain_nodes"
}

func (SequentialNodeModel) TableName() string {
	return "sequential_nodes"
}

func (ParallelNodeModel) TableName() string {
	return "parallel_nodes"
}

func (ParallelNodeChildModel) TableName() string {
	return "parallel_node_children"
}

func (IfElseNodeModel) TableName() string {
	return "if_else_nodes"
}

func (RepeatNodeModel) TableName() string {
	return "repeat_nodes"
}

func (ForEachNodeModel) TableName() string {
	return "for_each_nodes"
}

func (AtomicEffectModel) TableName() string {
	return "atomic_effects"
}

func (TargetSelectorModel) TableName() string {
	return "target_selectors"
}

func (TargetFilterModel) TableName() string {
	return "target_filters"
}

func (TargetFilterTraitModel) TableName() string {
	return "target_filter_traits"
}

func (ConditionModel) TableName() string {
	return "conditions"
}

// ValidateEffectChainNode 効果チェーンノードが対応する具体テーブルにレコードが存在するか検証
// 書き込み時はエラーを返し、読み込み時はwarningを返す
func ValidateEffectChainNode(db *gorm.DB, nodeID uint, nodeType string, isWrite bool) error {
	var exists bool
	var err error

	switch nodeType {
	case "THEN":
		var count int64
		err = db.Model(&SequentialNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "AND":
		var count int64
		err = db.Model(&ParallelNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "IF_ELSE":
		var count int64
		err = db.Model(&IfElseNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "REPEAT":
		var count int64
		err = db.Model(&RepeatNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "FOREACH":
		var count int64
		err = db.Model(&ForEachNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	default:
		return fmt.Errorf("unknown node type: %s", nodeType)
	}

	if err != nil {
		return fmt.Errorf("failed to validate effect chain node: %w", err)
	}

	if !exists {
		if isWrite {
			return fmt.Errorf("effect chain node (id=%d, type=%s) must have corresponding concrete table record", nodeID, nodeType)
		}
		// 読み込み時はwarning（ロガーに出力する想定）
		// ここではエラーを返さず、呼び出し側でwarningを処理
		return fmt.Errorf("WARNING: effect chain node (id=%d, type=%s) does not have corresponding concrete table record", nodeID, nodeType)
	}

	return nil
}

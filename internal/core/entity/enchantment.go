package entity

// ========================================
// エンチャント(永続効果)
// ゲーム中で発生する永続効果を表す構造体を定義する
// ========================================

// Enchantment エンチャント
type Enchantment struct {
	ID          string            `json:"id"`          // エンチャントID
	Name        string            `json:"name"`        // エンチャント名
	Description string            `json:"description"` // 説明文
	Effect      *EffectDefinition `json:"effect"`      // 効果定義（必須）

	// 持続期間管理
	Duration  *int `json:"duration,omitempty"`   // 持続ターン数（nilの場合は無期限）
	TurnsLeft *int `json:"turns_left,omitempty"` // 残りターン数

	// 発動タイミング
	// 永続効果は特定のタイミングで発動する
	Timing EffectTiming `json:"timing"` // 発動タイミング（TurnStart, TurnEndなど）

	// 対象管理
	OwnerID    string                `json:"owner_id"`            // 所有者のプレイヤーID
	TargetID   *string               `json:"target_id,omitempty"` // 対象ID（プレイヤーまたはユニット、nilの場合はオーナー自身）
	TargetType EnchantmentTargetType `json:"target_type"`         // 対象タイプ

	// メタデータ
	CreatedTurn  int     `json:"created_turn"`             // 作成されたターン
	SourceCardID *string `json:"source_card_id,omitempty"` // 元となったカードID（オプション）
}

// EnchantmentTargetType エンチャントの対象タイプ
type EnchantmentTargetType string

const (
	EnchantmentTargetPlayer EnchantmentTargetType = "Player" // プレイヤー
	EnchantmentTargetUnit   EnchantmentTargetType = "Unit"   // ユニット
	EnchantmentTargetField  EnchantmentTargetType = "Field"  // フィールド全体
)

// IsExpired エンチャントが期限切れか確認
func (e *Enchantment) IsExpired() bool {
	if e.TurnsLeft == nil {
		return false
	}
	return *e.TurnsLeft <= 0
}

// DecrementTurn ターン数を減らす
func (e *Enchantment) DecrementTurn() {
	if e.TurnsLeft != nil {
		*e.TurnsLeft--
	}
}

// ShouldTrigger 指定されたタイミングで発動すべきか確認
func (e *Enchantment) ShouldTrigger(timing EffectTiming) bool {
	return e.Timing == timing && !e.IsExpired()
}

// GetTargetPlayerID 対象のプレイヤーIDを取得
func (e *Enchantment) GetTargetPlayerID() string {
	if e.TargetType == EnchantmentTargetPlayer && e.TargetID != nil {
		return *e.TargetID
	}
	// デフォルトはオーナー
	return e.OwnerID
}

// Clone エンチャントのコピーを作成（ターン管理用）
func (e *Enchantment) Clone() *Enchantment {
	clone := *e
	if e.TurnsLeft != nil {
		turnsLeft := *e.TurnsLeft
		clone.TurnsLeft = &turnsLeft
	}
	if e.Duration != nil {
		duration := *e.Duration
		clone.Duration = &duration
	}
	if e.TargetID != nil {
		targetID := *e.TargetID
		clone.TargetID = &targetID
	}
	if e.SourceCardID != nil {
		sourceCardID := *e.SourceCardID
		clone.SourceCardID = &sourceCardID
	}
	return &clone
}

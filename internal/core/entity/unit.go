package entity

// ========================================
// ユニット
// 盤面に配置されたユニットの状態と能力を定義
// 設計方針:
// - 戦闘ダメージと効果ダメージを区別
// - 特性に応じた動的な能力変化に対応
// - 攻撃回数を管理（Windfury対応）
// ========================================

// Unit 盤面に配置されたユニット
type Unit struct {
	CardID           string  `json:"card_id"`
	InstanceID       string  `json:"instance_id"` // 盤面上での一意なID
	Name             string  `json:"name"`
	Cost             int     `json:"cost"`
	Attack           int     `json:"attack"`
	Defense          int     `json:"defense"`
	CurrentDefense   int     `json:"current_defense"`    // 現在の守備力
	Traits           []Trait `json:"traits"`             // 特性リスト
	Effect           string  `json:"effect"`             // 表示用効果テキスト
	AttacksRemaining int     `json:"attacks_remaining"`  // 残り攻撃回数
	SummonedThisTurn bool    `json:"summoned_this_turn"` // 召喚酔い判定用
	OwnerID          string  `json:"owner_id"`
}

// ========================================
// 特性判定
// ========================================

// HasTrait 指定された特性を持っているか確認
func (u *Unit) HasTrait(trait Trait) bool {
	return HasTrait(u.Traits, trait)
}

// HasAnyTrait いずれかの特性を持っているか確認
func (u *Unit) HasAnyTrait(traits []Trait) bool {
	return HasAnyTrait(u.Traits, traits)
}

// HasAllTraits すべての特性を持っているか確認
func (u *Unit) HasAllTraits(traits []Trait) bool {
	return HasAllTraits(u.Traits, traits)
}

// ========================================
// 特性操作
// ========================================

// AddTrait 特性を追加（重複チェック付き）
func (u *Unit) AddTrait(trait Trait) {
	u.Traits = AddTrait(u.Traits, trait)
}

// RemoveTrait 特性を削除
func (u *Unit) RemoveTrait(trait Trait) {
	u.Traits = RemoveTrait(u.Traits, trait)
}

// ========================================
// 攻撃判定
// ========================================

// CanAttack 攻撃可能か確認
func (u *Unit) CanAttack() bool {
	// 攻撃回数が残っていない場合は攻撃不可
	if u.AttacksRemaining <= 0 {
		return false
	}

	// Rush特性またはCharge特性を持つ場合は召喚酔いしない
	if u.HasTrait(TraitRush) || u.HasTrait(TraitCharge) {
		return true
	}

	// 召喚したターンは攻撃できない（召喚酔い）
	return !u.SummonedThisTurn
}

// CanDirectAttack プレイヤーに直接攻撃できるか確認
func (u *Unit) CanDirectAttack() bool {
	// Rush特性を持つ場合のみリーダーに直接攻撃可能
	// Charge特性の場合はユニットにのみ攻撃可能なので不可
	return u.HasTrait(TraitRush)
}

// HasAttacksRemaining 攻撃回数が残っているか確認
func (u *Unit) HasAttacksRemaining() bool {
	return u.AttacksRemaining > 0
}

// GetMaxAttacksPerTurn 1ターンの最大攻撃回数を取得
func (u *Unit) GetMaxAttacksPerTurn() int {
	// Windfury特性を持つ場合は2回攻撃可能
	if u.HasTrait(TraitWindfury) {
		return 2
	}
	return 1
}

// UseAttack 攻撃を行ったことを記録
func (u *Unit) UseAttack() {
	if u.AttacksRemaining > 0 {
		u.AttacksRemaining--
	}
}

// ========================================
// ダメージと回復
// ========================================

// TakeDamage ダメージを受ける
// isEffect: true = 効果ダメージ、false = 戦闘ダメージ
// 戻り値: ユニットが破壊されたかどうか
func (u *Unit) TakeDamage(amount int, isEffect bool) bool {
	// EffectShield特性: 効果ダメージを防ぐ
	if isEffect && u.HasTrait(TraitEffectShield) {
		return false
	}

	u.CurrentDefense -= amount
	return u.CurrentDefense <= 0
}

// Heal 守備力を回復
func (u *Unit) Heal(amount int) {
	u.CurrentDefense += amount
	if u.CurrentDefense > u.Defense {
		u.CurrentDefense = u.Defense
	}
}

// ========================================
// 対象指定判定
// ========================================

// CanBeTargeted スペル・効果の対象にできるか確認
func (u *Unit) CanBeTargeted() bool {
	// Untargetable特性を持つ場合は対象にできない
	return !u.HasTrait(TraitUntargetable)
}

// CanBeTargetedBySpell スペルの対象にできるか確認
func (u *Unit) CanBeTargetedBySpell() bool {
	return u.CanBeTargeted()
}

// CanBeTargetedByEffect 効果の対象にできるか確認
func (u *Unit) CanBeTargetedByEffect() bool {
	return u.CanBeTargeted()
}

// ========================================
// 状態判定
// ========================================

// IsDestroyed 破壊されているか確認
func (u *Unit) IsDestroyed() bool {
	return u.CurrentDefense <= 0
}

// IsHealthy 満タンのHPか確認
func (u *Unit) IsHealthy() bool {
	return u.CurrentDefense >= u.Defense
}

// IsDamaged ダメージを受けているか確認
func (u *Unit) IsDamaged() bool {
	return u.CurrentDefense < u.Defense
}

// GetRemainingDefense 残り守備力を取得
func (u *Unit) GetRemainingDefense() int {
	return u.CurrentDefense
}

// GetMissingDefense 失われた守備力を取得
func (u *Unit) GetMissingDefense() int {
	missing := u.Defense - u.CurrentDefense
	if missing < 0 {
		return 0
	}
	return missing
}

// ========================================
// ターン管理
// ========================================

// ResetForNewTurn 新しいターンのためにユニットの状態をリセット
func (u *Unit) ResetForNewTurn() {
	// 召喚酔いを解除
	if u.SummonedThisTurn {
		u.SummonedThisTurn = false
	}

	// 攻撃回数をリセット
	u.AttacksRemaining = u.GetMaxAttacksPerTurn()
}

// InitializeOnSummon 召喚時の初期化
func (u *Unit) InitializeOnSummon() {
	u.SummonedThisTurn = true

	// Rush特性またはCharge特性を持つ場合は召喚直後でも攻撃可能
	if u.HasTrait(TraitRush) || u.HasTrait(TraitCharge) {
		u.AttacksRemaining = u.GetMaxAttacksPerTurn()
	} else {
		u.AttacksRemaining = 0
	}
}

// ========================================
// 属性変更
// ========================================

// ModifyAttack 攻撃力を変更（バフ/デバフ）
func (u *Unit) ModifyAttack(amount int) {
	u.Attack += amount
	if u.Attack < 0 {
		u.Attack = 0
	}
}

// ModifyDefense 守備力を変更（バフ/デバフ）
func (u *Unit) ModifyDefense(amount int) {
	u.Defense += amount
	if u.Defense < 0 {
		u.Defense = 0
	}

	// 現在の守備力も調整
	u.CurrentDefense += amount
	if u.CurrentDefense < 0 {
		u.CurrentDefense = 0
	}
}

// SetAttack 攻撃力を設定
func (u *Unit) SetAttack(value int) {
	if value < 0 {
		value = 0
	}
	u.Attack = value
}

// SetDefense 守備力を設定
func (u *Unit) SetDefense(value int) {
	if value < 0 {
		value = 0
	}
	u.Defense = value

	// 現在の守備力が最大値を超えないように調整
	if u.CurrentDefense > u.Defense {
		u.CurrentDefense = u.Defense
	}
}

// ========================================
// 情報取得
// ========================================

// GetAttack 攻撃力を取得
func (u *Unit) GetAttack() int {
	return u.Attack
}

// GetDefense 守備力を取得
func (u *Unit) GetDefense() int {
	return u.Defense
}

// GetCurrentDefense 現在の守備力を取得
func (u *Unit) GetCurrentDefense() int {
	return u.CurrentDefense
}

// GetOwnerID 所有者IDを取得
func (u *Unit) GetOwnerID() string {
	return u.OwnerID
}

// GetInstanceID インスタンスIDを取得
func (u *Unit) GetInstanceID() string {
	return u.InstanceID
}

// GetName 名前を取得
func (u *Unit) GetName() string {
	return u.Name
}

// Clone ユニットのコピーを作成
func (u *Unit) Clone() Unit {
	traits := make([]Trait, len(u.Traits))
	copy(traits, u.Traits)

	return Unit{
		CardID:           u.CardID,
		InstanceID:       u.InstanceID,
		Name:             u.Name,
		Cost:             u.Cost,
		Attack:           u.Attack,
		Defense:          u.Defense,
		CurrentDefense:   u.CurrentDefense,
		Traits:           traits,
		Effect:           u.Effect,
		AttacksRemaining: u.AttacksRemaining,
		SummonedThisTurn: u.SummonedThisTurn,
		OwnerID:          u.OwnerID,
	}
}

package entity

import "time"

// ========================================
// イベント
// ゲーム内で発生するアクションと状態変化を記録
// 設計方針:
// - イベントは不変のデータ構造
// - プレイヤーアクションとその結果の両方を記録
// - カテゴリで区別し、将来的にCommand-Event分離に移行可能
// ========================================

// イベント種別
type EventType string

const (
	// アクションイベント（プレイヤーの入力）
	EventTypeAttack     EventType = "ATTACK"      // 攻撃アクション
	EventTypeSummonUnit EventType = "SUMMON_UNIT" // ユニット召喚アクション
	EventTypeUseSpell   EventType = "USE_SPELL"   // スペル使用アクション
	EventTypeMulligan   EventType = "MULLIGAN"    // マリガンアクション
	EventTypeEndTurn    EventType = "END_TURN"    // ターン終了アクション

	// 状態変更イベント（ゲーム状態の変化）
	EventTypeDamage       EventType = "DAMAGE"        // ダメージ発生
	EventTypeHeal         EventType = "HEAL"          // 回復発生
	EventTypeBuffDebuff   EventType = "BUFF_DEBUFF"   // バフ/デバフ適用
	EventTypeDraw         EventType = "DRAW"          // カードドロー
	EventTypeUnitSummoned EventType = "UNIT_SUMMONED" // ユニット召喚完了
	EventTypeDestroy      EventType = "DESTROY"       // ユニット破壊
	EventTypeDiscard      EventType = "DISCARD"       // カード破棄
	EventTypeEnchantment  EventType = "ENCHANTMENT"   // エンチャント適用/解除

	// ゲームフローイベント（ゲームの進行）
	EventTypeTurnStart   EventType = "TURN_START"   // ターン開始
	EventTypeTurnEnd     EventType = "TURN_END"     // ターン終了
	EventTypePhaseChange EventType = "PHASE_CHANGE" // フェーズ変更
	EventTypeGameOver    EventType = "GAME_OVER"    // ゲーム終了
)

// イベントカテゴリ
type EventCategory string

const (
	EventCategoryAction      EventCategory = "ACTION"       // プレイヤーアクション
	EventCategoryStateChange EventCategory = "STATE_CHANGE" // 状態変化
	EventCategoryGameFlow    EventCategory = "GAME_FLOW"    // ゲーム進行
)

// 回復タイプ
type HealType string

const (
	HealTypeHP      HealType = "HP"      // HP回復
	HealTypeAttack  HealType = "ATTACK"  // 攻撃力回復
	HealTypeDefense HealType = "DEFENSE" // 防御力回復
)

// 修正属性（バフ/デバフの対象属性）
type ModifierAttribute string

const (
	ModifierAttributeAttack  ModifierAttribute = "ATTACK"  // 攻撃力
	ModifierAttributeDefense ModifierAttribute = "DEFENSE" // 防御力
	ModifierAttributeCost    ModifierAttribute = "COST"    // コスト
	ModifierAttributeMaxHP   ModifierAttribute = "MAX_HP"  // 最大HP
)

// 破壊理由
type DestroyReason string

const (
	DestroyReasonCombat DestroyReason = "COMBAT" // 戦闘による破壊
	DestroyReasonEffect DestroyReason = "EFFECT" // 効果による破壊
	DestroyReasonDamage DestroyReason = "DAMAGE" // ダメージによる破壊
)

// ゲーム終了理由
type GameOverReason string

const (
	GameOverReasonPlayerDefeated GameOverReason = "PLAYER_DEFEATED" // プレイヤーのHP0
	GameOverReasonDeckOut        GameOverReason = "DECK_OUT"        // デッキ切れ
	GameOverReasonSurrender      GameOverReason = "SURRENDER"       // 降参
	GameOverReasonTimeout        GameOverReason = "TIMEOUT"         // タイムアウト
	GameOverReasonDraw           GameOverReason = "DRAW"            // 引き分け
)

// ゲームフェーズ
type GamePhase string

const (
	GamePhaseTurnStart    GamePhase = "TURN_START"    // ターン開始
	GamePhaseDraw         GamePhase = "DRAW"          // ドローフェーズ
	GamePhaseResourceGain GamePhase = "RESOURCE_GAIN" // リソース増加フェーズ
	GamePhaseMain         GamePhase = "MAIN"          // メインフェーズ
	GamePhaseTurnEnd      GamePhase = "TURN_END"      // ターン終了フェーズ
)

// イベントの基底インターフェース
type Event interface {
	Type() EventType         // イベント種別
	Category() EventCategory // イベントカテゴリ
	Timestamp() time.Time    // 発生時刻
	GameID() string          // ゲームID
	Turn() int               // ターン番号
}

// イベントの基底構造体
type baseEvent struct {
	eventType EventType
	timestamp time.Time
	gameID    string
	turn      int
}

func (e *baseEvent) Type() EventType {
	return e.eventType
}

func (e *baseEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *baseEvent) GameID() string {
	return e.gameID
}

func (e *baseEvent) Turn() int {
	return e.turn
}

// Category イベントカテゴリを取得
func (e *baseEvent) Category() EventCategory {
	switch e.eventType {
	case EventTypeAttack, EventTypeSummonUnit, EventTypeUseSpell, EventTypeMulligan, EventTypeEndTurn:
		return EventCategoryAction
	case EventTypeDamage, EventTypeHeal, EventTypeBuffDebuff, EventTypeDraw,
		EventTypeUnitSummoned, EventTypeDestroy, EventTypeDiscard, EventTypeEnchantment:
		return EventCategoryStateChange
	case EventTypeTurnStart, EventTypeTurnEnd, EventTypePhaseChange, EventTypeGameOver:
		return EventCategoryGameFlow
	default:
		return EventCategoryStateChange
	}
}

// ========================================
// アクションイベント（プレイヤーの入力）
// ========================================

// 攻撃アクション
type AttackAction struct {
	baseEvent
	AttackerID string  `json:"attacker_id"`         // 攻撃ユニットのInstanceID
	TargetID   *string `json:"target_id,omitempty"` // 対象ユニットのInstanceID（nilの場合はプレイヤー直接攻撃）
	PlayerID   string  `json:"player_id"`           // 攻撃プレイヤーのID
}

// NewAttackAction 攻撃アクションを作成
func NewAttackAction(gameID string, turn int, playerID, attackerID string, targetID *string) *AttackAction {
	return &AttackAction{
		baseEvent: baseEvent{
			eventType: EventTypeAttack,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		AttackerID: attackerID,
		TargetID:   targetID,
		PlayerID:   playerID,
	}
}

// IsDirectAttack プレイヤーへの直接攻撃か判定
func (a *AttackAction) IsDirectAttack() bool {
	return a.TargetID == nil
}

// ユニット召喚アクション
type SummonUnitAction struct {
	baseEvent
	PlayerID string `json:"player_id"`
	CardID   string `json:"card_id"`
}

// NewSummonUnitAction ユニット召喚アクションを作成
func NewSummonUnitAction(gameID string, turn int, playerID, cardID string) *SummonUnitAction {
	return &SummonUnitAction{
		baseEvent: baseEvent{
			eventType: EventTypeSummonUnit,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID: playerID,
		CardID:   cardID,
	}
}

// スペル使用アクション
type UseSpellAction struct {
	baseEvent
	PlayerID string  `json:"player_id"`
	CardID   string  `json:"card_id"`
	TargetID *string `json:"target_id,omitempty"` // 対象（nilの場合は対象不要なスペル）
}

// NewUseSpellAction スペル使用アクションを作成
func NewUseSpellAction(gameID string, turn int, playerID, cardID string, targetID *string) *UseSpellAction {
	return &UseSpellAction{
		baseEvent: baseEvent{
			eventType: EventTypeUseSpell,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID: playerID,
		CardID:   cardID,
		TargetID: targetID,
	}
}

// HasTarget 対象が指定されているか判定
func (u *UseSpellAction) HasTarget() bool {
	return u.TargetID != nil
}

// マリガンアクション
type MulliganAction struct {
	baseEvent
	PlayerID string   `json:"player_id"`
	CardIDs  []string `json:"card_ids"` // デッキに戻すカードIDのリスト
}

// NewMulliganAction マリガンアクションを作成
func NewMulliganAction(gameID string, turn int, playerID string, cardIDs []string) *MulliganAction {
	return &MulliganAction{
		baseEvent: baseEvent{
			eventType: EventTypeMulligan,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID: playerID,
		CardIDs:  cardIDs,
	}
}

// GetCardCount マリガンするカード枚数を取得
func (m *MulliganAction) GetCardCount() int {
	return len(m.CardIDs)
}

// ========================================
// 状態変更イベント（ゲーム状態の変化）
// ========================================

// ダメージイベント
type DamageEvent struct {
	baseEvent
	SourceID string `json:"source_id"` // ダメージ源のID
	TargetID string `json:"target_id"` // ダメージ対象のID
	Amount   int    `json:"amount"`    // ダメージ量
	IsEffect bool   `json:"is_effect"` // 効果ダメージか（戦闘ダメージでないか）
	IsDirect bool   `json:"is_direct"` // プレイヤーへの直接ダメージか
}

// NewDamageEvent ダメージイベントを作成
func NewDamageEvent(gameID string, turn int, sourceID, targetID string, amount int, isEffect, isDirect bool) *DamageEvent {
	return &DamageEvent{
		baseEvent: baseEvent{
			eventType: EventTypeDamage,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		SourceID: sourceID,
		TargetID: targetID,
		Amount:   amount,
		IsEffect: isEffect,
		IsDirect: isDirect,
	}
}

// IsCombatDamage 戦闘ダメージか判定
func (d *DamageEvent) IsCombatDamage() bool {
	return !d.IsEffect
}

// 回復イベント
type HealEvent struct {
	baseEvent
	SourceID string   `json:"source_id"` // 回復源のID
	TargetID string   `json:"target_id"` // 回復対象のID
	Amount   int      `json:"amount"`    // 回復量
	HealType HealType `json:"heal_type"` // 回復タイプ
}

// NewHealEvent 回復イベントを作成
func NewHealEvent(gameID string, turn int, sourceID, targetID string, amount int, healType HealType) *HealEvent {
	return &HealEvent{
		baseEvent: baseEvent{
			eventType: EventTypeHeal,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		SourceID: sourceID,
		TargetID: targetID,
		Amount:   amount,
		HealType: healType,
	}
}

// バフ/デバフイベント
type BuffDebuffEvent struct {
	baseEvent
	SourceID  string            `json:"source_id"`          // バフ/デバフ源のID
	TargetID  string            `json:"target_id"`          // 対象のID
	Attribute ModifierAttribute `json:"attribute"`          // 修正属性
	Amount    int               `json:"amount"`             // 変更量（正の値はバフ、負の値はデバフ）
	Duration  *int              `json:"duration,omitempty"` // 持続ターン数（nilの場合は永続）
}

// NewBuffDebuffEvent バフ/デバフイベントを作成
func NewBuffDebuffEvent(gameID string, turn int, sourceID, targetID string, attribute ModifierAttribute, amount int, duration *int) *BuffDebuffEvent {
	return &BuffDebuffEvent{
		baseEvent: baseEvent{
			eventType: EventTypeBuffDebuff,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		SourceID:  sourceID,
		TargetID:  targetID,
		Attribute: attribute,
		Amount:    amount,
		Duration:  duration,
	}
}

// IsBuff バフか判定（デバフでないか）
func (b *BuffDebuffEvent) IsBuff() bool {
	return b.Amount > 0
}

// IsDebuff デバフか判定
func (b *BuffDebuffEvent) IsDebuff() bool {
	return b.Amount < 0
}

// IsPermanent 永続効果か判定
func (b *BuffDebuffEvent) IsPermanent() bool {
	return b.Duration == nil
}

// ドローイベント
type DrawEvent struct {
	baseEvent
	PlayerID  string   `json:"player_id"`
	CardIDs   []string `json:"card_ids"`   // 引いたカードのIDリスト
	CardCount int      `json:"card_count"` // 引いた枚数
}

// NewDrawEvent ドローイベントを作成
func NewDrawEvent(gameID string, turn int, playerID string, cardIDs []string) *DrawEvent {
	return &DrawEvent{
		baseEvent: baseEvent{
			eventType: EventTypeDraw,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID:  playerID,
		CardIDs:   cardIDs,
		CardCount: len(cardIDs),
	}
}

// ユニット召喚完了イベント
type UnitSummonedEvent struct {
	baseEvent
	PlayerID   string `json:"player_id"`
	CardID     string `json:"card_id"`
	InstanceID string `json:"instance_id"` // 召喚されたユニットのInstanceID
}

// NewUnitSummonedEvent ユニット召喚完了イベントを作成
func NewUnitSummonedEvent(gameID string, turn int, playerID, cardID, instanceID string) *UnitSummonedEvent {
	return &UnitSummonedEvent{
		baseEvent: baseEvent{
			eventType: EventTypeUnitSummoned,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID:   playerID,
		CardID:     cardID,
		InstanceID: instanceID,
	}
}

// ユニット破壊イベント
type DestroyEvent struct {
	baseEvent
	UnitID  string        `json:"unit_id"`  // 破壊されたユニットのInstanceID
	OwnerID string        `json:"owner_id"` // 所有者のプレイヤーID
	Reason  DestroyReason `json:"reason"`   // 破壊理由
}

// NewDestroyEvent ユニット破壊イベントを作成
func NewDestroyEvent(gameID string, turn int, unitID, ownerID string, reason DestroyReason) *DestroyEvent {
	return &DestroyEvent{
		baseEvent: baseEvent{
			eventType: EventTypeDestroy,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		UnitID:  unitID,
		OwnerID: ownerID,
		Reason:  reason,
	}
}

// ========================================
// ゲームフローイベント（ゲームの進行）
// ========================================

// ターン開始イベント
type TurnStartEvent struct {
	baseEvent
	PlayerID string    `json:"player_id"` // ターンを開始するプレイヤーのID
	Phase    GamePhase `json:"phase"`     // フェーズ
}

// NewTurnStartEvent ターン開始イベントを作成
func NewTurnStartEvent(gameID string, turn int, playerID string, phase GamePhase) *TurnStartEvent {
	return &TurnStartEvent{
		baseEvent: baseEvent{
			eventType: EventTypeTurnStart,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID: playerID,
		Phase:    phase,
	}
}

// ターン終了イベント
type TurnEndEvent struct {
	baseEvent
	PlayerID string `json:"player_id"` // ターンを終了するプレイヤーのID
}

// NewTurnEndEvent ターン終了イベントを作成
func NewTurnEndEvent(gameID string, turn int, playerID string) *TurnEndEvent {
	return &TurnEndEvent{
		baseEvent: baseEvent{
			eventType: EventTypeTurnEnd,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		PlayerID: playerID,
	}
}

// ゲーム終了イベント
type GameOverEvent struct {
	baseEvent
	WinnerID *string        `json:"winner_id,omitempty"` // 勝者のID（nilの場合は引き分け）
	Reason   GameOverReason `json:"reason"`              // 終了理由
}

// NewGameOverEvent ゲーム終了イベントを作成
func NewGameOverEvent(gameID string, turn int, winnerID *string, reason GameOverReason) *GameOverEvent {
	return &GameOverEvent{
		baseEvent: baseEvent{
			eventType: EventTypeGameOver,
			timestamp: time.Now(),
			gameID:    gameID,
			turn:      turn,
		},
		WinnerID: winnerID,
		Reason:   reason,
	}
}

// IsDraw 引き分けか判定
func (g *GameOverEvent) IsDraw() bool {
	return g.WinnerID == nil
}

// HasWinner 勝者がいるか判定
func (g *GameOverEvent) HasWinner() bool {
	return g.WinnerID != nil
}

// ========================================
// ゲームログ
// ========================================

// ゲームログのエントリ
type GameLogEntry struct {
	Turn      int       `json:"turn"`
	Phase     string    `json:"phase"`
	PlayerID  string    `json:"player_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

package event

import (
	"fmt"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// イベントハンドラー実装
// port.EventHandlerの各種実装
// 設計方針:
// - 用途別に複数のハンドラーを提供
// - 組み合わせ可能な設計（Composite, Filtered）
// - 拡張しやすい構造
// ========================================

// コンパイル時チェック: 各ハンドラーがport.EventHandlerを実装していることを確認
var (
	_ port.EventHandler = (*LoggingHandler)(nil)
	_ port.EventHandler = (*GameLogHandler)(nil)
	_ port.EventHandler = (*CompositeHandler)(nil)
	_ port.EventHandler = (*FilteredHandler)(nil)
)

// ========================================
// LoggingHandler
// ========================================

// イベントをログに記録するハンドラー
type LoggingHandler struct {
	logger port.Logger
}

// ログハンドラーを作成
func NewLoggingHandler(logger port.Logger) *LoggingHandler {
	return &LoggingHandler{
		logger: logger,
	}
}

// イベントを処理（ログに記録）
func (h *LoggingHandler) Handle(event entity.Event) error {
	if h.logger == nil {
		return nil
	}

	// イベントタイプに応じてログを出力
	switch e := event.(type) {
	case *entity.DamageEvent:
		h.logger.Info("Damage: %s -> %s (%d)", e.SourceID, e.TargetID, e.Amount)
	case *entity.HealEvent:
		h.logger.Info("Heal: %s -> %s (%d)", e.SourceID, e.TargetID, e.Amount)
	case *entity.DrawEvent:
		h.logger.Info("Draw: %s drew %d cards", e.PlayerID, e.CardCount)
	case *entity.UnitSummonedEvent:
		h.logger.Info("Summon: %s summoned %s", e.PlayerID, e.InstanceID)
	case *entity.DestroyEvent:
		h.logger.Info("Destroy: %s destroyed (reason: %s)", e.UnitID, e.Reason)
	default:
		h.logger.Info("Event: %s", event.Type())
	}

	return nil
}

// すべてのイベントタイプを処理可能
func (h *LoggingHandler) CanHandle(eventType entity.EventType) bool {
	return true
}

// ========================================
// GameLogHandler
// ========================================

// イベントをゲームログに記録するハンドラー
type GameLogHandler struct {
	addLog func(playerID, action, details string) // ログ追加関数
}

// ゲームログハンドラーを作成
func NewGameLogHandler(addLog func(playerID, action, details string)) *GameLogHandler {
	return &GameLogHandler{
		addLog: addLog,
	}
}

// イベントを処理（ゲームログに記録）
func (h *GameLogHandler) Handle(event entity.Event) error {
	if h.addLog == nil {
		return nil
	}

	// イベントタイプに応じてログエントリを作成
	var playerID, action, details string

	switch e := event.(type) {
	case *entity.DamageEvent:
		playerID = e.SourceID
		action = "ダメージ"
		details = fmt.Sprintf("%s に %d ダメージ", e.TargetID, e.Amount)
	case *entity.HealEvent:
		playerID = e.SourceID
		action = "回復"
		details = fmt.Sprintf("%s を %d 回復", e.TargetID, e.Amount)
	case *entity.DrawEvent:
		playerID = e.PlayerID
		action = "ドロー"
		details = fmt.Sprintf("%d 枚ドロー", e.CardCount)
	case *entity.UnitSummonedEvent:
		playerID = e.PlayerID
		action = "召喚"
		details = fmt.Sprintf("%s を召喚", e.InstanceID)
	case *entity.DestroyEvent:
		playerID = e.OwnerID
		action = "破壊"
		details = fmt.Sprintf("%s が破壊された", e.UnitID)
	default:
		return nil // ログに記録しないイベント
	}

	// ゲームログに追加
	h.addLog(playerID, action, details)

	return nil
}

// 特定のイベントタイプのみ処理
func (h *GameLogHandler) CanHandle(eventType entity.EventType) bool {
	switch eventType {
	case entity.EventTypeDamage, entity.EventTypeHeal, entity.EventTypeDraw,
		entity.EventTypeUnitSummoned, entity.EventTypeDestroy:
		return true
	default:
		return false
	}
}

// ========================================
// CompositeHandler
// ========================================

// 複数のハンドラーを組み合わせるハンドラー
type CompositeHandler struct {
	handlers []port.EventHandler
}

// 複合ハンドラーを作成
func NewCompositeHandler(handlers ...port.EventHandler) *CompositeHandler {
	return &CompositeHandler{
		handlers: handlers,
	}
}

// すべてのハンドラーにイベントを渡す
func (h *CompositeHandler) Handle(event entity.Event) error {
	for _, handler := range h.handlers {
		if handler.CanHandle(event.Type()) {
			if err := handler.Handle(event); err != nil {
				return fmt.Errorf("handler error: %w", err)
			}
		}
	}
	return nil
}

// いずれかのハンドラーが処理可能かどうか
func (h *CompositeHandler) CanHandle(eventType entity.EventType) bool {
	for _, handler := range h.handlers {
		if handler.CanHandle(eventType) {
			return true
		}
	}
	return false
}

// ========================================
// FilteredHandler
// ========================================

// 特定の条件でイベントをフィルタリングするハンドラー
type FilteredHandler struct {
	handler port.EventHandler
	filter  func(entity.Event) bool
}

// フィルタリングハンドラーを作成
func NewFilteredHandler(handler port.EventHandler, filter func(entity.Event) bool) *FilteredHandler {
	return &FilteredHandler{
		handler: handler,
		filter:  filter,
	}
}

// フィルタを通過したイベントのみ処理
func (h *FilteredHandler) Handle(event entity.Event) error {
	if h.filter != nil && !h.filter(event) {
		return nil // フィルタを通過しない場合は処理しない
	}

	if h.handler != nil {
		return h.handler.Handle(event)
	}
	return nil
}

// 元のハンドラーが処理可能かどうか
func (h *FilteredHandler) CanHandle(eventType entity.EventType) bool {
	if h.handler != nil {
		return h.handler.CanHandle(eventType)
	}
	return false
}

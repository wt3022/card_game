package event

import (
	"fmt"
	"sync"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// イベントバス実装
// port.EventBusの実装
// 設計方針:
// - In-Memory実装（シンプル）
// - 同期/非同期の発行をサポート
// - スレッドセーフ
// - 将来的に他の実装（Redis, Kafka）に置き換え可能
// ========================================

// イベントバスの実装
type Bus struct {
	mu       sync.RWMutex
	handlers map[entity.EventType][]port.EventHandler
}

// コンパイル時チェック: Busがport.EventBusを実装していることを確認
var _ port.EventBus = (*Bus)(nil)

// 新しいイベントバスを作成
func NewBus() port.EventBus {
	return &Bus{
		handlers: make(map[entity.EventType][]port.EventHandler),
	}
}

// ========================================
// イベント発行
// ========================================

// イベントを非同期で発行
func (b *Bus) Publish(event entity.Event) error {
	if event == nil {
		return fmt.Errorf("イベントがnilです")
	}

	b.mu.RLock()
	handlers := b.handlers[event.Type()]
	b.mu.RUnlock()

	// 非同期でハンドラーを実行
	for _, handler := range handlers {
		if handler.CanHandle(event.Type()) {
			go func(h port.EventHandler, e entity.Event) {
				if err := h.Handle(e); err != nil {
					// エラーログは呼び出し側で処理することを想定
					// ここではエラーを無視（ハンドラーのエラーでイベント処理を停止しない）
				}
			}(handler, event)
		}
	}

	return nil
}

// イベントを同期的に発行（すべてのハンドラーが処理完了するまで待つ）
func (b *Bus) PublishSync(event entity.Event) error {
	if event == nil {
		return fmt.Errorf("イベントがnilです")
	}

	b.mu.RLock()
	handlers := b.handlers[event.Type()]
	b.mu.RUnlock()

	// 同期的にハンドラーを実行
	for _, handler := range handlers {
		if handler.CanHandle(event.Type()) {
			if err := handler.Handle(event); err != nil {
				return fmt.Errorf("イベント %s のハンドラーエラー: %w", event.Type(), err)
			}
		}
	}

	return nil
}

// ========================================
// ハンドラー管理
// ========================================

// イベントタイプを購読
func (b *Bus) Subscribe(eventType entity.EventType, handler port.EventHandler) {
	if handler == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	// 既に登録されているかチェック
	for _, h := range handlers {
		if h == handler {
			return // 既に登録されている
		}
	}

	b.handlers[eventType] = append(handlers, handler)
}

// イベントタイプの購読を解除
func (b *Bus) Unsubscribe(eventType entity.EventType, handler port.EventHandler) {
	if handler == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

// ========================================
// ユーティリティ
// ========================================

// すべてのハンドラーをクリア
func (b *Bus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers = make(map[entity.EventType][]port.EventHandler)
}

// イベントタイプのハンドラー数を取得
func (b *Bus) GetHandlerCount(eventType entity.EventType) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.handlers[eventType])
}

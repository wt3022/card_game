package port

import "card_game/internal/core/entity"

// ========================================
// イベント処理インターフェース
// イベントの発行と購読を定義
// 設計方針:
// - イベント駆動アーキテクチャをサポート
// - ハンドラーパターンで拡張性を確保
// - 同期/非同期の発行を選択可能
// ========================================

// ========================================
// イベントハンドラー
// ========================================

// イベントを処理するハンドラー
type EventHandler interface {
	// イベントを処理
	Handle(event entity.Event) error

	// 指定されたイベントタイプを処理できるか判定
	CanHandle(eventType entity.EventType) bool
}

// ========================================
// イベントバス
// ========================================

// イベントの発行と購読を管理
type EventBus interface {
	// イベントを非同期で発行
	Publish(event entity.Event) error

	// イベントタイプを購読
	Subscribe(eventType entity.EventType, handler EventHandler)

	// イベントタイプの購読を解除
	Unsubscribe(eventType entity.EventType, handler EventHandler)

	// 同期的にイベントを発行（すべてのハンドラーが処理完了するまで待つ）
	PublishSync(event entity.Event) error
}

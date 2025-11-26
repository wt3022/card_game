package usecase

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// ユニット破壊通知システム
// Observerパターンの実装
// 設計方針:
// - 複数のハンドラーを登録可能
// - usecase層間の疎結合を維持
// ========================================

// DestructionNotifier ユニット破壊の通知を管理
type DestructionNotifier struct {
	handlers []port.UnitDestructionHandler
}

// NewDestructionNotifier 通知システムを作成
func NewDestructionNotifier() *DestructionNotifier {
	return &DestructionNotifier{
		handlers: make([]port.UnitDestructionHandler, 0),
	}
}

// RegisterHandler ハンドラーを登録
func (n *DestructionNotifier) RegisterHandler(handler port.UnitDestructionHandler) {
	n.handlers = append(n.handlers, handler)
}

// NotifyUnitDestruction ユニット破壊を通知
func (n *DestructionNotifier) NotifyUnitDestruction(unit *entity.Unit, owner *entity.Player) error {
	for _, handler := range n.handlers {
		if err := handler.OnUnitDestroyed(unit, owner); err != nil {
			// エラーが発生しても他のハンドラーは実行する
			// ログに記録するなどの処理が必要
			continue
		}
	}
	return nil
}

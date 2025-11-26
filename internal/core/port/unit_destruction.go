package port

import "card_game/internal/core/entity"

// ========================================
// ユニット破壊通知インターフェース
// ユニット破壊時の処理を抽象化
// 設計方針:
// - Observer パターンの適用
// - usecase層間の疎結合を維持
// ========================================

// UnitDestructionHandler ユニット破壊時の処理を行うハンドラー
type UnitDestructionHandler interface {
	// ユニット破壊時に呼ばれる
	OnUnitDestroyed(unit *entity.Unit, owner *entity.Player) error
}

// UnitDestructionNotifier ユニット破壊の通知を管理
type UnitDestructionNotifier interface {
	// ハンドラーを登録
	RegisterHandler(handler UnitDestructionHandler)

	// ユニット破壊を通知
	NotifyUnitDestruction(unit *entity.Unit, owner *entity.Player) error
}

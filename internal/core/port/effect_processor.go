package port

import "card_game/internal/core/entity"

// ========================================
// 効果処理インターフェース
// 効果処理への依存を抽象化
// 設計方針:
// - 依存性逆転の原則（DIP）を実現
// - usecase層間の疎結合を維持
// ========================================

// EffectProcessor 効果処理の抽象インターフェース
type EffectProcessor interface {
	// 特定のタイミングで発動する効果を処理
	ProcessTimingEffects(card *entity.Card, timing entity.EffectTiming, playerID string, targetID *string) error

	// 効果定義を処理
	ProcessEffectDefinition(def *entity.EffectDefinition, sourcePlayerID string, targetID *string) error
}

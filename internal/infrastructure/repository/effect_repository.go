package repository

import (
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence/model"
	"fmt"

	"gorm.io/gorm"
)

// EffectRepository 効果チェーンノードを扱うリポジトリ（将来実装用の例）
type EffectRepository struct {
	db     *gorm.DB
	logger port.Logger
}

// SaveEffectChainNode 効果チェーンノードを保存（書き込み時はバリデーション）
func (r *EffectRepository) SaveEffectChainNode(node *model.EffectChainNodeModel) error {
	// 書き込み前にバリデーション（エラーを返す）
	if err := ValidateEffectChainNode(r.db, r.logger, node.ID, node.Type, true); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 保存処理
	if err := r.db.Save(node).Error; err != nil {
		return fmt.Errorf("failed to save effect chain node: %w", err)
	}

	return nil
}

// LoadEffectChainNode 効果チェーンノードを読み込み（読み込み時はwarning）
func (r *EffectRepository) LoadEffectChainNode(nodeID uint) (*model.EffectChainNodeModel, error) {
	var node model.EffectChainNodeModel
	if err := r.db.First(&node, nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load effect chain node: %w", err)
	}

	// 読み込み時にバリデーション（warningを出力）
	if err := ValidateEffectChainNode(r.db, r.logger, node.ID, node.Type, false); err != nil {
		// warningは既にログに出力されているので、エラーは返さない
		// ただし、データが不完全である可能性があることを呼び出し側に知らせる
	}

	return &node, nil
}

package repository

import (
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence/model"
	"fmt"

	"gorm.io/gorm"
)

// ValidateEffectChainNode 効果チェーンノードが対応する具体テーブルにレコードが存在するか検証
// isWrite: true=書き込み時（エラーを返す）、false=読み込み時（warningを返す）
func ValidateEffectChainNode(db *gorm.DB, logger port.Logger, nodeID uint, nodeType string, isWrite bool) error {
	var exists bool
	var err error

	switch nodeType {
	case "THEN":
		var count int64
		err = db.Model(&model.SequentialNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "AND":
		var count int64
		err = db.Model(&model.ParallelNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "IF_ELSE":
		var count int64
		err = db.Model(&model.IfElseNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "REPEAT":
		var count int64
		err = db.Model(&model.RepeatNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	case "FOREACH":
		var count int64
		err = db.Model(&model.ForEachNodeModel{}).Where("node_id = ?", nodeID).Count(&count).Error
		exists = count > 0
	default:
		return fmt.Errorf("unknown node type: %s", nodeType)
	}

	if err != nil {
		return fmt.Errorf("failed to validate effect chain node: %w", err)
	}

	if !exists {
		if isWrite {
			return fmt.Errorf("effect chain node (id=%d, type=%s) must have corresponding concrete table record", nodeID, nodeType)
		}
		// 読み込み時はwarningをログに出力
		if logger != nil {
			logger.Info("効果チェーンノード (id=%d, type=%s) に対応する具体テーブルレコードがありません。無視します", nodeID, nodeType)
		}
	}

	return nil
}

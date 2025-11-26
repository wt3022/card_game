package unit

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// 効果無効化（サイレンス）
func ExecuteSilenceUnit(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	if len(targets) == 0 {
		return nil
	}

	for _, t := range targets {
		unit, ok := t.(*entity.Unit)
		if !ok {
			continue
		}

		// 効果盾チェック：効果を受けない
		if unit.HasTrait(entity.TraitEffectShield) {
			game.AddLog(sourcePlayer.ID, "サイレンス無効", fmt.Sprintf("%s は効果盾によりサイレンスを受けない", unit.Name))
			continue
		}

		// 効果無効化: ユニットの効果テキストをクリアし、特性を全て除去する
		unit.Effect = ""
		unit.Traits = []entity.Trait{}
		game.AddLog(sourcePlayer.ID, "サイレンス", fmt.Sprintf("%s の効果を無効化した", unit.Name))
	}
	return nil
}

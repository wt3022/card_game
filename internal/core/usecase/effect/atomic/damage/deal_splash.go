package damage

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
)

// 範囲ダメージ（複数ターゲットに一括ダメージ）
func ExecuteDealSplash(effect *entity.AtomicEffect, sourcePlayer, opponent *entity.Player, targets []any, game port.GameStateReader) error {
	// 安全のため、ユニットはインスタンスIDで処理する。
	// これにより、処理中にフィールドが変更されても正しく現在のユニットを取得してダメージを適用できる。

	// まずプレイヤーへのダメージ対象を決定
	hitPlayer := false
	if effect.Target.Type == entity.EffectTargetEnemies {
		// 敵全体指定ではプレイヤーも対象に含める設計（過去の実装に合わせる）
		hitPlayer = true
	}

	// ユニットのIDリストを作る
	type unitRef struct {
		OwnerID    string
		InstanceID string
		Name       string
	}

	unitRefs := []unitRef{}
	for _, target := range targets {
		if u, ok := target.(*entity.Unit); ok {
			unitRefs = append(unitRefs, unitRef{OwnerID: u.OwnerID, InstanceID: u.InstanceID, Name: u.Name})
		}
	}

	// まずプレイヤーへダメージ（先に与えることでログ順を保つ）
	if hitPlayer {
		opponent.TakeDamage(effect.Value)
		game.AddLog(sourcePlayer.ID, "範囲ダメージ", fmt.Sprintf("%s に %d ダメージ (範囲)", opponent.Name, effect.Value))
	}

	// ユニット群へダメージ（IDで再取得して適用）
	for _, ref := range unitRefs {
		owner := game.GetPlayerByID(ref.OwnerID)
		if owner == nil {
			continue
		}

		u := owner.GetUnitByInstanceID(ref.InstanceID)
		if u == nil {
			// 既に破壊済みまたは存在しない
			continue
		}

		// 詳細ログ: 適用前の耐久
		game.AddLog(sourcePlayer.ID, "範囲ダメージ(前)", fmt.Sprintf("%s の現在守備 %d に %d ダメージを適用", u.Name, u.CurrentDefense, effect.Value))

		destroyed := u.TakeDamage(effect.Value, true)

		// ダメージログ
		game.AddLog(sourcePlayer.ID, "範囲ダメージ", fmt.Sprintf("%s に %d ダメージ (範囲)", u.Name, effect.Value))

		if destroyed {
			owner.RemoveUnitFromField(u.InstanceID)
			game.AddLog(owner.ID, "ユニット破壊", fmt.Sprintf("%s が破壊された (範囲)", u.Name))
		}
	}

	return nil
}

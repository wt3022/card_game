package unit

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
	"math/rand"
	"time"
)

// ExecuteSummonUnit: 効果によるユニット召喚（トークン生成や指定IDのユニットを場に出す）
// Parameters例: {"unit_proto": *entity.Unit, "unit_id": string, ...}
func ExecuteSummonUnit(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	if effect.Parameters == nil {
		return entity.NewErrEffectNotImplemented("SUMMON_UNIT(no parameters)")
	}

	// 1. プリセットユニット（unit_proto）で召喚
	if proto, ok := effect.Parameters["unit_proto"].(*entity.Unit); ok && proto != nil {
		unit := proto.Clone()
		unit.InstanceID = generateInstanceID(unit.CardID)
		unit.OwnerID = sourcePlayer.ID
		unit.InitializeOnSummon()
		sourcePlayer.Field = append(sourcePlayer.Field, unit)
		game.AddLog(sourcePlayer.ID, "召喚", fmt.Sprintf("%s を召喚", unit.Name))
		return nil
	}

	// 2. unit_id でマスターデータから生成（未実装: ゲーム側で雛形取得APIが必要）
	if unitID, ok := effect.Parameters["unit_id"].(string); ok && unitID != "" {
		return entity.NewErrEffectNotImplemented("SUMMON_UNIT(unit_id指定/雛形生成未実装)")
	}

	// 3. unit_name で生成（省略）
	if unitName, ok := effect.Parameters["unit_name"].(string); ok && unitName != "" {
		return entity.NewErrEffectNotImplemented("SUMMON_UNIT(unit_name指定)")
	}

	return entity.NewErrEffectNotImplemented("SUMMON_UNIT(unknown parameters)")
}

// generateInstanceID: ユニットの一意なインスタンスIDを生成（CardID+乱数+時刻）
func generateInstanceID(cardID string) string {
	return fmt.Sprintf("%s-%d-%d", cardID, rand.Intn(1000000), time.Now().UnixNano())
}

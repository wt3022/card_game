package unit

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"fmt"
	"math/rand"
	"time"
)

// ExecuteSummonUnit: 効果によるユニット召喚（トークン生成や指定IDのユニットを場に出す）
// Parameters例:
//   - "unit_proto": *entity.Unit - プリセットユニット
//   - "card_id": string - カードIDからユニットを生成
//   - "count": int - 召喚する数（デフォルト1）
//   - "attack": int - 攻撃力を上書き
//   - "defense": int - 防御力を上書き
//   - "traits": []entity.Trait - 特性を追加
func ExecuteSummonUnit(effect *entity.AtomicEffect, sourcePlayer *entity.Player, targets []any, game port.GameStateReader) error {
	if effect.Parameters == nil {
		return entity.NewErrEffectNotImplemented("SUMMON_UNIT(no parameters)")
	}

	// 召喚数を取得（デフォルト1）
	count := 1
	if c, ok := effect.Parameters["count"].(int); ok && c > 0 {
		count = c
	}

	// フィールドの上限チェック（7体）
	if len(sourcePlayer.Field)+count > 7 {
		return fmt.Errorf("フィールドがいっぱいです")
	}

	// 1. プリセットユニット（unit_proto）で召喚
	if proto, ok := effect.Parameters["unit_proto"].(*entity.Unit); ok && proto != nil {
		return summonFromProto(proto, sourcePlayer, count, effect.Parameters, game)
	}

	// 2. card_id でカードデータからユニットを生成
	if cardID, ok := effect.Parameters["card_id"].(string); ok && cardID != "" {
		return summonFromCardID(cardID, sourcePlayer, count, effect.Parameters, game)
	}

	// 3. カスタムユニットを生成（名前、攻撃力、防御力指定）
	if name, ok := effect.Parameters["name"].(string); ok && name != "" {
		return summonCustomUnit(name, sourcePlayer, count, effect.Parameters, game)
	}

	return entity.NewErrEffectNotImplemented("SUMMON_UNIT(unknown parameters)")
}

// summonFromProto プリセットユニットから召喚
func summonFromProto(proto *entity.Unit, player *entity.Player, count int, params map[string]interface{}, game port.GameStateReader) error {
	for i := 0; i < count; i++ {
		unit := proto.Clone()
		unit.InstanceID = generateInstanceID(unit.CardID)
		unit.OwnerID = player.ID

		// パラメータで上書き
		applyUnitModifications(&unit, params)

		unit.InitializeOnSummon()
		player.Field = append(player.Field, unit)
		game.AddLog(player.ID, "召喚", fmt.Sprintf("%s を召喚", unit.Name))
	}
	return nil
}

// summonFromCardID カードIDからユニットを生成して召喚
func summonFromCardID(cardID string, player *entity.Player, count int, params map[string]interface{}, game port.GameStateReader) error {
	// カードリポジトリからカード情報を取得
	// Note: GameStateReaderにカードリポジトリへのアクセスが必要
	// 現状ではゲームステートからカード情報を取得できないため、デフォルト値でユニットを生成

	for i := 0; i < count; i++ {
		unit := entity.Unit{
			CardID:         cardID,
			InstanceID:     generateInstanceID(cardID),
			Name:           cardID, // 名前はパラメータで上書き可能
			Cost:           0,
			Attack:         1, // デフォルト値
			Defense:        1,
			CurrentDefense: 1,
			Traits:         []entity.Trait{},
			Effect:         "", // トークンはデフォルトで効果なし
			OwnerID:        player.ID,
		}

		// パラメータで上書き
		applyUnitModifications(&unit, params)

		unit.InitializeOnSummon()
		player.Field = append(player.Field, unit)
		game.AddLog(player.ID, "召喚", fmt.Sprintf("%s を召喚", unit.Name))
	}
	return nil
}

// summonCustomUnit カスタムユニットを召喚
func summonCustomUnit(name string, player *entity.Player, count int, params map[string]interface{}, game port.GameStateReader) error {
	for i := 0; i < count; i++ {
		unit := entity.Unit{
			CardID:         "token-" + name,
			InstanceID:     generateInstanceID("token-" + name),
			Name:           name,
			Cost:           0,
			Attack:         1,
			Defense:        1,
			CurrentDefense: 1,
			Traits:         []entity.Trait{},
			Effect:         "", // トークンはデフォルトで効果なし
			OwnerID:        player.ID,
		}

		// パラメータで上書き
		applyUnitModifications(&unit, params)

		unit.InitializeOnSummon()
		player.Field = append(player.Field, unit)
		game.AddLog(player.ID, "召喚", fmt.Sprintf("%s を召喚", unit.Name))
	}
	return nil
}

// applyUnitModifications パラメータを使ってユニットの統計を修正
func applyUnitModifications(unit *entity.Unit, params map[string]interface{}) {
	// 名前の上書き
	if name, ok := params["name"].(string); ok && name != "" {
		unit.Name = name
	}

	// 攻撃力の上書き
	if attack, ok := params["attack"].(int); ok {
		unit.Attack = attack
	}

	// 防御力の上書き
	if defense, ok := params["defense"].(int); ok {
		unit.Defense = defense
		unit.CurrentDefense = defense
	}

	// コストの上書き
	if cost, ok := params["cost"].(int); ok {
		unit.Cost = cost
	}

	// 特性の追加
	if traits, ok := params["traits"].([]entity.Trait); ok {
		for _, trait := range traits {
			unit.AddTrait(trait)
		}
	}

	// 効果テキストの上書き
	if effect, ok := params["effect"].(string); ok && effect != "" {
		unit.Effect = effect
	}
}

// generateInstanceID: ユニットの一意なインスタンスIDを生成（CardID+乱数+時刻）
func generateInstanceID(cardID string) string {
	// time.Now().UnixNano()で十分に一意性が確保されるが、念のため乱数も追加
	return fmt.Sprintf("%s-%d-%d", cardID, rand.Intn(1000000), time.Now().UnixNano())
}

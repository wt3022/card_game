package usecase

import (
	"fmt"
	"strings"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// カードプレイ処理
// ユニット召喚とスペル使用のロジック
// 設計方針:
// - カード操作を一箇所に集約
// - マナ返却などのロールバック処理を統一
// - 対象指定の検証を実装
// ========================================

// ========================================
// ユニット召喚
// ========================================

// ユニットを召喚
func executeSummonUnit(state port.GameState, playerID string, cardID string) (*entity.Card, error) {
	player := state.GetPlayerByID(playerID)
	if player == nil {
		return nil, entity.NewErrNotFound("player", playerID)
	}

	// 手札からカードをプレイ
	card, err := player.PlayCardFromHand(cardID)
	if err != nil {
		return nil, err
	}

	// ユニットカードか確認
	if !card.IsUnit() {
		// マナを返却してエラー
		player.AddMana(card.Cost)
		return nil, entity.NewErrInvalidCardType("unit", string(card.Type))
	}

	// ユニットインスタンスIDを生成
	instanceID := generateUnitInstanceID(player, state.GetCurrentTurn())

	// ユニットを召喚
	err = player.SummonUnit(*card, instanceID)
	if err != nil {
		// エラーの場合はマナを返却
		player.AddMana(card.Cost)
		return nil, err
	}

	return card, nil
}

// ユニットインスタンスIDを生成
func generateUnitInstanceID(player *entity.Player, currentTurn int) string {
	return fmt.Sprintf("unit-%s-%d-%d", player.GetID(), currentTurn, player.GetFieldSize())
}

// ========================================
// スペル使用
// ========================================

// スペルを使用
func executeUseSpell(state port.GameState, playerID string, cardID string, targetID *string) (*entity.Card, error) {
	player := state.GetPlayerByID(playerID)
	if player == nil {
		return nil, entity.NewErrNotFound("player", playerID)
	}

	// カードの取得（まだ手札から削除しない）
	card := findCardInHand(player, cardID)
	if card == nil {
		return nil, entity.NewErrNotFound("card", cardID)
	}

	// スペルカードか確認
	if !card.IsSpell() {
		return nil, entity.NewErrInvalidCardType("spell", string(card.Type))
	}

	// 対象が必要な効果かチェック
	if err := validateSpellTarget(card, targetID); err != nil {
		return nil, err
	}

	// マナを消費し、カードを手札から削除
	playedCard, err := player.PlayCardFromHand(cardID)
	if err != nil {
		return nil, err
	}

	// スペルを使用（墓地に送る）
	player.UseSpell(*playedCard)
	return playedCard, nil
}

// 手札からカードを検索
func findCardInHand(player *entity.Player, cardID string) *entity.Card {
	for i := range player.Hand {
		if player.Hand[i].ID == cardID {
			return &player.Hand[i]
		}
	}
	return nil
}

// スペルの対象指定を検証
func validateSpellTarget(card *entity.Card, targetID *string) error {
	requiresTarget := false

	// CardEffectから対象指定が必要か判定
	if card.CardEffect != nil {
		for _, def := range card.CardEffect.Definitions {
			if def.RequireTarget {
				requiresTarget = true
				break
			}
		}
	} else {
		// CardEffectが未定義の場合、効果テキストから推測
		requiresTarget = containsTargetKeywords(card.Effect)
	}

	// 対象が必要なのに指定されていない
	if requiresTarget && targetID == nil {
		return entity.NewErrInvalidAction("use_spell", fmt.Sprintf("spell '%s' requires a target", card.Name))
	}

	return nil
}

// 効果テキストに対象指定キーワードが含まれるか確認
func containsTargetKeywords(effectText string) bool {
	lowerEffect := strings.ToLower(effectText)
	keywords := []string{
		"to a unit",
		"to target",
		"choose a",
		"select a",
		"target unit",
		"target enemy",
		"target ally",
	}

	for _, keyword := range keywords {
		if strings.Contains(lowerEffect, keyword) {
			return true
		}
	}
	return false
}

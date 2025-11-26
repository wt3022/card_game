package game

import (
	"fmt"

	"card_game/internal/core/entity"
)

// ========================================
// ゲームフェイズ処理
// 各フェイズの実行ロジック
// 設計方針:
// - 各フェイズを独立した関数として定義
// - ロガーを活用して情報出力
// - エンチャント処理を含む
// ========================================

// ターン開始フェイズを実行
func (g *State) ExecuteTurnStartPhase(player *entity.Player) {
	g.CurrentPhase = entity.GamePhaseTurnStart
	g.logger.Info("--- ターン開始フェイズ ---")

	// プレイヤーのターンフラグをリセット
	player.ResetTurnFlags()

	// エンチャントの処理
	g.ProcessEnchantments(player)

	g.AddLog(player.ID, "ターン開始フェイズ", "")
}

// ドローフェイズを実行
func (g *State) ExecuteDrawPhase(player *entity.Player) {
	g.CurrentPhase = entity.GamePhaseDraw
	g.logger.Info("--- ドローフェイズ ---")

	// 先攻1ターン目はドローをスキップ
	if player.IsFirstTurn && g.CurrentTurn == 1 {
		g.logger.Info("%s: 先攻1ターン目のためドローをスキップ", player.Name)
		player.IsFirstTurn = false
		g.AddLog(player.ID, "ドローフェイズ", "先攻1ターン目のためスキップ")
		return
	}

	// カードをドロー
	card, err := player.DrawCard()
	if err != nil {
		// デッキ切れの場合、即敗北
		g.logger.Info("%s: デッキ切れ！敗北", player.Name)
		g.IsGameOver = true
		opponent := g.GetOpponent(player.ID)
		g.WinnerID = &opponent.ID
		g.AddLog(player.ID, "ドローフェイズ", "デッキ切れにより敗北")
		return
	}

	g.logger.Info("%s: 「%s」をドロー (手札: %d枚, デッキ: %d枚)", player.Name, card.Name, len(player.Hand), len(player.Deck))
	g.AddLog(player.ID, "ドローフェイズ", fmt.Sprintf("%s をドロー", card.Name))
}

// リソース増加フェイズを実行
func (g *State) ExecuteResourceGainPhase(player *entity.Player) {
	g.CurrentPhase = entity.GamePhaseResourceGain
	g.logger.Info("--- リソース増加フェイズ ---")

	beforeMana := player.CurrentTurnMana

	// 最大回復マナを増加
	player.IncrementMaxRecoveryMana()

	// マナ回復: CurrentTurnMana = BeforeTurnMana + MaxRecoveryMana
	player.RecoverMana()

	g.logger.Info("%s: マナ %d → %d (前ターン残り: %d, 回復: +%d, 最大回復マナ: %d)",
		player.Name, beforeMana, player.CurrentTurnMana, beforeMana, player.CurrentRecoveryMana, player.CurrentRecoveryMana)
	g.AddLog(player.ID, "リソース増加フェイズ",
		fmt.Sprintf("マナ回復 %d→%d (最大回復マナ: %d)", beforeMana, player.CurrentTurnMana, player.CurrentRecoveryMana))
}

// メインフェイズを実行
func (g *State) ExecuteMainPhase(player *entity.Player) {
	g.CurrentPhase = entity.GamePhaseMain
	g.logger.Info("--- メインフェイズ ---")

	// このフェイズではプレイヤーの入力を待つ
	// サンプル実装ではAIや自動処理を行う
}

// ターン終了フェイズを実行
func (g *State) ExecuteTurnEndPhase(player *entity.Player) {
	g.CurrentPhase = entity.GamePhaseTurnEnd
	g.logger.Info("--- ターン終了フェイズ ---")

	g.logger.Info("%s: ターン終了 (残りマナ: %d は次のターンに持ち越し)",
		player.Name, player.CurrentTurnMana)

	g.AddLog(player.ID, "ターン終了フェイズ", "")
}

// ========================================
// エンチャント処理
// ========================================

// エンチャントの処理
func (g *State) ProcessEnchantments(player *entity.Player) {
	if len(g.Enchantments) == 0 {
		return
	}

	g.logger.Info("エンチャントの処理:")
	remainingEnchantments := []entity.Enchantment{}

	for _, enchantment := range g.Enchantments {
		if enchantment.OwnerID == player.ID {
			g.logger.Info("  - %s: %s", enchantment.Name, enchantment.Description)

			// 期限のあるエンチャントのターン数を減らす
			enchantment.DecrementTurn()
			if !enchantment.IsExpired() {
				remainingEnchantments = append(remainingEnchantments, enchantment)
			} else {
				g.logger.Info("    → %s が終了", enchantment.Name)
			}
		} else {
			remainingEnchantments = append(remainingEnchantments, enchantment)
		}
	}

	g.Enchantments = remainingEnchantments
}

// ========================================
// ヘルパー関数
// ========================================

// ファティーグダメージを計算
func calculateFatigueDamage(currentTurn int) int {
	return 1 + (currentTurn / 10) // ターンが進むほどダメージ増加
}

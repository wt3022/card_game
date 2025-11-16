package game

import (
	"card_game/internal/core/entity"
)

// ========================================
// 勝利条件判定
// ゲームの勝敗を判定
// 設計方針:
// - HP判定
// - デッキ切れ判定
// - 勝者の決定
// ========================================

// 勝利条件をチェック
func (g *State) CheckVictoryConditions() {
	// プレイヤー1のHPチェック
	if g.Player1.IsDefeated() {
		g.IsGameOver = true
		g.WinnerID = &g.Player2.ID
		g.logger.Info("%s が HP 0 で敗北。%s の勝利！", g.Player1.Name, g.Player2.Name)
		return
	}

	// プレイヤー2のHPチェック
	if g.Player2.IsDefeated() {
		g.IsGameOver = true
		g.WinnerID = &g.Player1.ID
		g.logger.Info("%s が HP 0 で敗北。%s の勝利！", g.Player2.Name, g.Player1.Name)
		return
	}

	// デッキ切れチェック
	if !g.Player1.HasCardsInDeck() {
		g.IsGameOver = true
		g.WinnerID = &g.Player2.ID
		g.logger.Info("%s がデッキ切れで敗北。%s の勝利！", g.Player1.Name, g.Player2.Name)
		return
	}

	if !g.Player2.HasCardsInDeck() {
		g.IsGameOver = true
		g.WinnerID = &g.Player1.ID
		g.logger.Info("%s がデッキ切れで敗北。%s の勝利！", g.Player2.Name, g.Player1.Name)
		return
	}
}

// 勝者を取得
func (g *State) GetWinner() *entity.Player {
	if g.WinnerID == nil {
		return nil
	}
	return g.GetPlayerByID(*g.WinnerID)
}

// ゲーム終了フラグを取得
func (g *State) IsOver() bool {
	return g.IsGameOver
}

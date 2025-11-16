package game

import (
	"fmt"
	"time"

	"card_game/internal/core/entity"
)

// ========================================
// ゲームログ管理
// ログの追加と取得
// 設計方針:
// - ログエントリの追加
// - ログの取得
// - ログの表示
// ========================================

// ゲームログを追加
func (g *State) AddLog(playerID string, action string, details string) {
	entry := entity.GameLogEntry{
		Turn:      g.CurrentTurn,
		Phase:     string(g.CurrentPhase),
		PlayerID:  playerID,
		Action:    action,
		Timestamp: time.Now(),
		Details:   details,
	}
	g.GameLog = append(g.GameLog, entry)
}

// 最近のログを取得
func (g *State) GetRecentLogs(count int) []entity.GameLogEntry {
	if count <= 0 || count > len(g.GameLog) {
		return g.GameLog
	}
	return g.GameLog[len(g.GameLog)-count:]
}

// ========================================
// 表示関数
// ========================================

// ゲームログを表示
func (g *State) PrintGameLog() {
	fmt.Println("\n=== ゲームログ ===")
	for _, entry := range g.GameLog {
		fmt.Printf("[T%d %s] %s: %s %s\n",
			entry.Turn, entry.Phase, entry.PlayerID, entry.Action, entry.Details)
	}
}

// ゲーム結果を表示
func (g *State) PrintGameResult() {
	if g.WinnerID != nil {
		winner := g.GetWinner()
		if winner != nil {
			fmt.Printf("\n勝者: %s\n", winner.Name)
		}
	} else if g.IsDraw {
		fmt.Println("\n引き分け")
	} else {
		fmt.Println("\nゲーム継続中")
	}
	fmt.Printf("総ターン数: %d\n", g.CurrentTurn)
}

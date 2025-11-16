package game

// ========================================
// ターン処理
// ターン全体のフロー制御
// 設計方針:
// - 各フェイズを順次実行
// - ゲーム終了判定
// - ターン交代
// ========================================

// ターン全体を実行
func (g *State) ExecuteTurn() {
	g.CurrentTurn++
	currentPlayer := g.GetCurrentPlayer()

	// 1. ターン開始フェイズ
	g.ExecuteTurnStartPhase(currentPlayer)
	if g.IsGameOver {
		return
	}

	// 2. ドローフェイズ
	g.ExecuteDrawPhase(currentPlayer)
	if g.IsGameOver {
		return
	}

	// 3. リソース増加フェイズ
	g.ExecuteResourceGainPhase(currentPlayer)

	// 4. メインフェイズ
	g.ExecuteMainPhase(currentPlayer)
	if g.IsGameOver {
		return
	}

	// 5. ターン終了フェイズ
	g.ExecuteTurnEndPhase(currentPlayer)

	// ターン交代
	g.SwitchTurn()
}

package port

import "card_game/internal/core/entity"

// ========================================
// ゲーム状態インターフェース
// ゲーム状態へのアクセスと操作を定義
// 設計方針:
// - 読み取り専用と変更可能を分離
// - 戦闘とターン管理を独立したインターフェースに
// - 依存性逆転の原則（DIP）を実現
// ========================================

// ========================================
// 読み取り専用アクセス
// ========================================

// ゲーム状態の読み取り専用アクセス
type GameStateReader interface {
	// プレイヤーをIDで取得
	GetPlayerByID(playerID string) *entity.Player

	// 相手プレイヤーを取得
	GetOpponent(playerID string) *entity.Player

	// ゲームログを追加
	AddLog(playerID, typ, msg string)

	// ロガーを取得
	GetLogger() Logger
}

// ========================================
// 戦闘状態読み取り
// ========================================

// 戦闘に必要な状態読み取り
type CombatStateReader interface {
	GameStateReader

	// 相手フィールドにGuardianユニットがいるか確認
	HasGuardianUnits(player *entity.Player) bool

	// Guardianユニットのリストを取得
	GetGuardianUnits(player *entity.Player) []string

	// 勝利条件をチェック
	CheckVictoryConditions()
}

// ========================================
// ターン制御
// ========================================

// ターン管理とフェイズ実行
type TurnController interface {
	// 現在のプレイヤーを取得
	GetCurrentPlayer() *entity.Player

	// 現在のターン番号を取得
	GetCurrentTurn() int

	// ターン番号をインクリメント
	IncrementCurrentTurn()

	// ターン開始フェイズを実行
	ExecuteTurnStartPhase(player *entity.Player)

	// ドローフェイズを実行
	ExecuteDrawPhase(player *entity.Player)

	// リソース増加フェイズを実行
	ExecuteResourceGainPhase(player *entity.Player)

	// ターン終了フェイズを実行
	ExecuteTurnEndPhase(player *entity.Player)

	// ターンを交代
	SwitchTurn()

	// ゲームが終了しているか確認
	IsOver() bool
}

// ========================================
// 統合インターフェース
// ========================================

// ゲーム状態の完全なインターフェース
// concrete な game.State はこのインターフェースを実装する
type GameState interface {
	GameStateReader
	CombatStateReader
	TurnController
}

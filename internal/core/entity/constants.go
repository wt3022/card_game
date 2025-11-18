package entity

import "time"

// ========================================
// ゲーム内で使用する定数を定義する
// ========================================

// ゲームの定数定義
const (
	// マナ関連
	MaxMana         = 10 // 最大マナ
	InitialMana     = 0  // 初期マナ
	MaxRecoveryMana = 10 // 最大回復マナ

	// プレイヤー関連
	InitialHP       = 20 // 初期HP
	InitialHandSize = 4  // 初期手札枚数

	// ターン関連
	DefaultTurnTime = 300 * time.Second // 初期持ち時間
)

package game

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// ゲーム状態
// ゲームの状態を管理
// 設計方針:
// - ゲーム状態の定義
// - プレイヤー操作
// - ターン管理
// - ユニット操作
// ========================================

// ゲーム状態
type State struct {
	GameID          string                `json:"game_id"`
	Player1         *entity.Player        `json:"player1"`
	Player2         *entity.Player        `json:"player2"`
	CurrentPlayerID string                `json:"current_player_id"`
	CurrentTurn     int                   `json:"current_turn"`
	CurrentPhase    entity.GamePhase      `json:"current_phase"`
	Enchantments    []entity.Enchantment  `json:"enchantments"`
	GameLog         []entity.GameLogEntry `json:"game_log"`
	IsGameOver      bool                  `json:"is_game_over"`
	WinnerID        *string               `json:"winner_id,omitempty"`
	IsDraw          bool                  `json:"is_draw"`
	logger          port.Logger           `json:"-"`
}

// NewState は新しいゲーム状態を作成
func NewState(gameID string, player1, player2 *entity.Player, logger port.Logger) *State {
	if logger == nil {
		logger = port.NewConsoleLogger()
	}
	return &State{
		GameID:          gameID,
		Player1:         player1,
		Player2:         player2,
		CurrentPlayerID: player1.ID,
		CurrentTurn:     1,
		CurrentPhase:    entity.GamePhaseTurnStart,
		Enchantments:    []entity.Enchantment{},
		GameLog:         []entity.GameLogEntry{},
		IsGameOver:      false,
		logger:          logger,
	}
}

// ========================================
// ロガー操作
// ========================================

// ロガーを取得
func (g *State) GetLogger() port.Logger {
	return g.logger
}

// ========================================
// プレイヤー操作
// ========================================

// 現在のプレイヤーを取得
func (g *State) GetCurrentPlayer() *entity.Player {
	if g.CurrentPlayerID == g.Player1.ID {
		return g.Player1
	}
	return g.Player2
}

// 相手プレイヤーを取得
func (g *State) GetOpponent(playerID string) *entity.Player {
	if playerID == g.Player1.ID {
		return g.Player2
	}
	return g.Player1
}

// プレイヤーIDでプレイヤーを取得
func (g *State) GetPlayerByID(playerID string) *entity.Player {
	if g.Player1.ID == playerID {
		return g.Player1
	}
	return g.Player2
}

// ========================================
// ターン管理
// ========================================

// ターンを交代
func (g *State) SwitchTurn() {
	if g.CurrentPlayerID == g.Player1.ID {
		g.CurrentPlayerID = g.Player2.ID
	} else {
		g.CurrentPlayerID = g.Player1.ID
	}
}

// 現在のターン番号を返す
func (g *State) GetCurrentTurn() int {
	return g.CurrentTurn
}

// ターン番号をインクリメント
func (g *State) IncrementCurrentTurn() {
	g.CurrentTurn++
}

// ========================================
// ユニット操作
// ========================================

// 攻撃可能なユニットを取得
func (g *State) GetAttackableUnits(playerID string) []*entity.Unit {
	player := g.GetPlayerByID(playerID)
	if player == nil {
		return nil
	}

	attackable := []*entity.Unit{}
	for i := range player.Field {
		unit := &player.Field[i]
		if unit.CanAttack() {
			attackable = append(attackable, unit)
		}
	}
	return attackable
}

// 相手フィールドにGuardianを持つユニットがいるか確認
func (g *State) HasGuardianUnits(player *entity.Player) bool {
	for i := range player.Field {
		if player.Field[i].HasTrait(entity.TraitGuardian) {
			return true
		}
	}
	return false
}

// Guardianを持つユニットのリストを取得
func (g *State) GetGuardianUnits(player *entity.Player) []string {
	guardians := []string{}
	for i := range player.Field {
		if player.Field[i].HasTrait(entity.TraitGuardian) {
			guardians = append(guardians, player.Field[i].InstanceID)
		}
	}
	return guardians
}

// ========================================
// アクション検証
// ========================================

// アクションの妥当性を検証
func (g *State) ValidateAction(playerID string) error {
	if g.IsGameOver {
		return entity.NewErrInvalidState("game_over", "ゲームは既に終了しています")
	}
	if g.CurrentPlayerID != playerID {
		return entity.NewErrNotYourTurn(playerID)
	}
	return nil
}

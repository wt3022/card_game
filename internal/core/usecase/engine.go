package usecase

import (
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/core/usecase/combat"
	"card_game/internal/core/usecase/effect"
)

// ========================================
// ユースケースエンジン
// ゲームロジックの統合窓口
// 設計方針:
// - 各サブシステムへの委譲
// - ビジネスロジックの orchestration
// - 外部から呼び出しやすい統一API
// ========================================

// ユースケース層のゲームエンジン
type Engine struct {
	State           port.GameState
	EffectProcessor *effect.Processor
}

// 新しいユースケースエンジンを作成
func NewEngine(state port.GameState) *Engine {
	return &Engine{
		State:           state,
		EffectProcessor: effect.NewProcessor(state),
	}
}

// ========================================
// カード操作
// ========================================

// ユニットを召喚
// 前提: 呼び出し側で ValidateAction を済ませること
func (e *Engine) SummonUnit(playerID string, cardID string) (*entity.Card, error) {
	return executeSummonUnit(e.State, playerID, cardID)
}

// スペルを使用
// 前提: 呼び出し側で ValidateAction を済ませること
func (e *Engine) UseSpell(playerID string, cardID string, targetID *string) (*entity.Card, error) {
	return executeUseSpell(e.State, playerID, cardID, targetID)
}

// ========================================
// 戦闘処理
// ========================================

// 攻撃を実行
// 前提: 呼び出し側で ValidateAction を済ませること
func (e *Engine) ExecuteAttack(action entity.AttackAction) (*entity.CombatResult, error) {
	return combat.ExecuteAttack(e.State, action)
}

// 攻撃可能なターゲットを取得
func (e *Engine) GetAttackableTargets(playerID string, attackerID string) ([]string, bool, error) {
	return combat.GetAttackableTargets(e.State, playerID, attackerID)
}

// ========================================
// ターン処理
// ========================================

func (e *Engine) StartTurn(playerID string) error {
	player := e.State.GetPlayerByID(playerID)
	if player == nil {
		return entity.NewErrNotFound("player", playerID)
	}

	// 1. 次プレイヤーの開始処理
	nextPlayer := e.State.GetCurrentPlayer()
	e.State.ExecuteTurnStartPhase(nextPlayer)

	// 2. ドローフェイズ(先行1ターン目はスキップ)
	if e.State.GetCurrentTurn() > 1 {
		e.State.ExecuteDrawPhase(nextPlayer)
	}

	// 3. リソース増加フェイズ
	e.State.ExecuteResourceGainPhase(nextPlayer)

	// 4. 勝利判定
	e.State.CheckVictoryConditions()

	return nil
}

// ターンを終了
func (e *Engine) EndTurn(playerID string) error {
	player := e.State.GetPlayerByID(playerID)
	if player == nil {
		return entity.NewErrNotFound("player", playerID)
	}

	// 1. ターン終了フェイズ
	e.State.ExecuteTurnEndPhase(player)

	// 2. ターン切り替え
	e.State.SwitchTurn()

	// 3. ターン番号をインクリメント
	e.State.IncrementCurrentTurn()

	return nil
}

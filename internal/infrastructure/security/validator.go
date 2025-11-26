package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"card_game/internal/core/usecase/game"
)

// ========================================
// ゲーム状態検証システム
// クライアント側での改ざんを検出
// ========================================

// StateValidator ゲーム状態の検証を行う
type StateValidator struct {
	stateHashes map[string]StateHash
	mu          sync.RWMutex
}

// StateHash ゲーム状態のハッシュ情報
type StateHash struct {
	Hash      string
	Timestamp time.Time
	TurnCount int
}

// NewStateValidator 新しいStateValidatorを作成
func NewStateValidator() *StateValidator {
	return &StateValidator{
		stateHashes: make(map[string]StateHash),
	}
}

// GenerateStateHash ゲーム状態からハッシュを生成
func GenerateStateHash(state *game.State) string {
	// ゲーム状態の重要な要素を結合
	data := fmt.Sprintf(
		"game:%s|turn:%d|phase:%s|currentPlayer:%s|p1hp:%d|p1mana:%d|p2hp:%d|p2mana:%d",
		state.GameID,
		state.CurrentTurn,
		state.CurrentPhase,
		state.CurrentPlayerID,
		state.Player1.HP,
		state.Player1.CurrentTurnMana,
		state.Player2.HP,
		state.Player2.CurrentTurnMana,
	)

	// 各プレイヤーの手札枚数とフィールドユニット数も含める
	data += fmt.Sprintf("|p1hand:%d|p1field:%d|p2hand:%d|p2field:%d",
		len(state.Player1.Hand),
		len(state.Player1.Field),
		len(state.Player2.Hand),
		len(state.Player2.Field))

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// RecordState ゲーム状態を記録
func (sv *StateValidator) RecordState(gameID string, state *game.State) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	hash := GenerateStateHash(state)
	sv.stateHashes[gameID] = StateHash{
		Hash:      hash,
		Timestamp: time.Now(),
		TurnCount: state.CurrentTurn,
	}
}

// ValidateState ゲーム状態が改ざんされていないか検証
func (sv *StateValidator) ValidateState(gameID string, state *game.State) error {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	recorded, exists := sv.stateHashes[gameID]
	if !exists {
		// 初回は記録のみ
		return nil
	}

	currentHash := GenerateStateHash(state)

	// ターン数が進んでいない場合、ハッシュは一致すべき
	if state.CurrentTurn == recorded.TurnCount && currentHash != recorded.Hash {
		return fmt.Errorf("game state tampered: hash mismatch for turn %d", state.CurrentTurn)
	}

	return nil
}

// CleanupOldStates 古い状態を削除
func (sv *StateValidator) CleanupOldStates(maxAge time.Duration) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	now := time.Now()
	for gameID, stateHash := range sv.stateHashes {
		if now.Sub(stateHash.Timestamp) > maxAge {
			delete(sv.stateHashes, gameID)
		}
	}
}

// ========================================
// チート検出システム
// 不正な操作パターンを検出
// ========================================

// CheatDetector チート検出を行う
type CheatDetector struct {
	suspiciousActions map[string][]SuspiciousAction
	mu                sync.Mutex
}

// SuspiciousAction 疑わしいアクション
type SuspiciousAction struct {
	Timestamp time.Time
	Action    string
	Reason    string
	Severity  int // 1-5 (1:低, 5:高)
}

// NewCheatDetector 新しいCheatDetectorを作成
func NewCheatDetector() *CheatDetector {
	return &CheatDetector{
		suspiciousActions: make(map[string][]SuspiciousAction),
	}
}

// DetectImpossibleTiming 不可能なタイミングでの操作を検出
func (cd *CheatDetector) DetectImpossibleTiming(playerID string, action string, lastActionTime time.Time) {
	now := time.Now()
	timeSinceLastAction := now.Sub(lastActionTime)

	// 人間には不可能な速さ（100ms未満）で連続操作
	if timeSinceLastAction < 100*time.Millisecond {
		cd.recordSuspiciousAction(playerID, SuspiciousAction{
			Timestamp: now,
			Action:    action,
			Reason:    fmt.Sprintf("Too fast action: %v since last action", timeSinceLastAction),
			Severity:  3,
		})
	}
}

// DetectImpossibleAction 不可能なアクションを検出
func (cd *CheatDetector) DetectImpossibleAction(playerID string, action string, reason string) {
	cd.recordSuspiciousAction(playerID, SuspiciousAction{
		Timestamp: time.Now(),
		Action:    action,
		Reason:    reason,
		Severity:  4,
	})
}

// recordSuspiciousAction 疑わしいアクションを記録
func (cd *CheatDetector) recordSuspiciousAction(playerID string, action SuspiciousAction) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	actions := cd.suspiciousActions[playerID]
	actions = append(actions, action)
	cd.suspiciousActions[playerID] = actions

	// 疑わしいアクションが閾値を超えたら警告
	if cd.calculateSuspicionScore(playerID) >= 10 {
		// ログに記録、管理者に通知など
		fmt.Printf("WARNING: Player %s has high suspicion score\n", playerID)
	}
}

// calculateSuspicionScore 疑わしいスコアを計算
func (cd *CheatDetector) calculateSuspicionScore(playerID string) int {
	actions := cd.suspiciousActions[playerID]
	score := 0

	// 過去5分間のアクションのみカウント
	cutoff := time.Now().Add(-5 * time.Minute)
	for _, action := range actions {
		if action.Timestamp.After(cutoff) {
			score += action.Severity
		}
	}

	return score
}

// GetSuspiciousPlayers 疑わしいプレイヤーのリストを取得
func (cd *CheatDetector) GetSuspiciousPlayers(threshold int) []string {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	var suspicious []string
	for playerID := range cd.suspiciousActions {
		if cd.calculateSuspicionScore(playerID) >= threshold {
			suspicious = append(suspicious, playerID)
		}
	}

	return suspicious
}

// ResetPlayer プレイヤーの記録をリセット
func (cd *CheatDetector) ResetPlayer(playerID string) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	delete(cd.suspiciousActions, playerID)
}

// Cleanup 古いアクション記録をクリーンアップ
func (cd *CheatDetector) Cleanup(maxAge time.Duration) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for playerID, actions := range cd.suspiciousActions {
		var valid []SuspiciousAction
		for _, action := range actions {
			if action.Timestamp.After(cutoff) {
				valid = append(valid, action)
			}
		}

		if len(valid) == 0 {
			delete(cd.suspiciousActions, playerID)
		} else {
			cd.suspiciousActions[playerID] = valid
		}
	}
}

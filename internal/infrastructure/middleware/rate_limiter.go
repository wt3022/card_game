package middleware

import (
	"fmt"
	"sync"
	"time"
)

// ========================================
// レート制限ミドルウェア
// 短時間での大量リクエストを防ぐ
// ========================================

// RateLimiter レート制限を管理する構造体
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	// アクションごとの制限設定
	limits map[string]RateLimit
}

// RateLimit レート制限の設定
type RateLimit struct {
	MaxRequests int           // 最大リクエスト数
	Window      time.Duration // 時間窓
}

// NewRateLimiter 新しいRateLimiterを作成
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limits: map[string]RateLimit{
			"PlayCard":        {MaxRequests: 10, Window: time.Second},    // カードプレイ: 1秒に10回まで
			"ExecuteAttack":   {MaxRequests: 10, Window: time.Second},    // 攻撃: 1秒に10回まで
			"EndTurn":         {MaxRequests: 5, Window: time.Second},     // ターン終了: 1秒に5回まで
			"PerformMulligan": {MaxRequests: 1, Window: 5 * time.Second}, // マリガン: 5秒に1回まで
			"default":         {MaxRequests: 30, Window: time.Second},    // デフォルト: 1秒に30回まで
		},
	}
}

// CheckLimit レート制限をチェック
func (rl *RateLimiter) CheckLimit(playerID string, action string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// アクションに対応する制限設定を取得
	limit, ok := rl.limits[action]
	if !ok {
		limit = rl.limits["default"]
	}

	key := fmt.Sprintf("%s:%s", playerID, action)
	now := time.Now()

	// 時間窓内のリクエストをフィルタリング
	var recentRequests []time.Time
	for _, t := range rl.requests[key] {
		if now.Sub(t) < limit.Window {
			recentRequests = append(recentRequests, t)
		}
	}

	// レート制限チェック
	if len(recentRequests) >= limit.MaxRequests {
		return fmt.Errorf("rate limit exceeded: %d requests in %v for action %s",
			len(recentRequests), limit.Window, action)
	}

	// 新しいリクエストを記録
	rl.requests[key] = append(recentRequests, now)

	return nil
}

// Reset 特定のプレイヤーのレート制限をリセット
func (rl *RateLimiter) Reset(playerID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// プレイヤーに関連するすべてのキーを削除
	for key := range rl.requests {
		if len(key) > len(playerID) && key[:len(playerID)] == playerID {
			delete(rl.requests, key)
		}
	}
}

// Cleanup 古いエントリを定期的にクリーンアップ
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	maxWindow := 10 * time.Second // 最大保持時間

	for key, timestamps := range rl.requests {
		var valid []time.Time
		for _, t := range timestamps {
			if now.Sub(t) < maxWindow {
				valid = append(valid, t)
			}
		}

		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

// StartCleanupRoutine クリーンアップルーチンを開始
func (rl *RateLimiter) StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			rl.Cleanup()
		}
	}()
}

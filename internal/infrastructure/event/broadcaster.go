package event

import (
	"sync"
	"time"

	"card_game/internal/core/usecase/game"
)

// ========================================
// イベントブロードキャスター
// ゲームイベントのリアルタイム配信
// 設計方針:
// - WebSocketやSSE用のイベント配信
// - ゲームセッションごとに購読管理
// - スレッドセーフ
// ========================================

// ゲームイベント
type GameEvent struct {
	GameID    string
	EventType string // "game_created", "card_played", "attack_executed", "turn_ended"
	Message   string
	PlayerID  string    // イベントを発生させたプレイヤーID
	Timestamp time.Time // イベント発生時刻
	State     *game.State
}

// ゲームイベントのブロードキャスト管理
type Broadcaster struct {
	mu      sync.RWMutex
	streams map[string][]chan *GameEvent
}

// 新しいBroadcasterを作成
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		streams: make(map[string][]chan *GameEvent),
	}
}

// ゲームイベントを購読
func (b *Broadcaster) Subscribe(gameID string) chan *GameEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan *GameEvent, 100)
	b.streams[gameID] = append(b.streams[gameID], ch)
	return ch
}

// ゲームイベントの購読を解除
func (b *Broadcaster) Unsubscribe(gameID string, ch chan *GameEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	streams := b.streams[gameID]
	for i, stream := range streams {
		if stream == ch {
			close(ch)
			b.streams[gameID] = append(streams[:i], streams[i+1:]...)
			break
		}
	}
}

// すべての購読者にイベントをブロードキャスト
func (b *Broadcaster) Broadcast(gameID string, event *GameEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.streams[gameID] {
		select {
		case ch <- event:
		default:
			// チャネルがブロックされている場合はスキップ
		}
	}
}

// ゲーム終了時のクリーンアップ
func (b *Broadcaster) CleanupGame(gameID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.streams[gameID] {
		close(ch)
	}
	delete(b.streams, gameID)
}

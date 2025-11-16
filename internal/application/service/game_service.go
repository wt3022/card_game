package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/core/usecase"
	"card_game/internal/core/usecase/effect"
	"card_game/internal/core/usecase/game"
	"card_game/internal/infrastructure/event"
)

// GameSession はゲームセッションを表す構造体
type GameSession struct {
	State            *game.State
	UsecaseEngine    *usecase.Engine
	EffectProcessor  *effect.Processor
	ConnectedPlayers map[string]bool // 接続中のプレイヤーを追跡
	mu               sync.RWMutex
}

// GameService はゲームセッションを管理するサービス
type GameService struct {
	mu               sync.RWMutex
	games            map[string]*GameSession
	eventBroadcaster *event.Broadcaster
	logger           port.Logger
}

// NewGameService は新しいGameServiceを作成
func NewGameService(logger port.Logger) *GameService {
	if logger == nil {
		logger = port.NewConsoleLogger()
	}
	return &GameService{
		games:            make(map[string]*GameSession),
		eventBroadcaster: event.NewBroadcaster(),
		logger:           logger,
	}
}

// CreateGame は新しいゲームセッションを作成
func (s *GameService) CreateGame(gameID string, player1ID, player1Name, player2ID, player2Name string, deck1, deck2 []entity.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[gameID]; exists {
		return entity.NewErrAlreadyExists("game", gameID)
	}

	// デッキをシャッフル
	player1Deck := make([]entity.Card, len(deck1))
	player2Deck := make([]entity.Card, len(deck2))
	copy(player1Deck, deck1)
	copy(player2Deck, deck2)

	rand.Shuffle(len(player1Deck), func(i, j int) {
		player1Deck[i], player1Deck[j] = player1Deck[j], player1Deck[i]
	})
	rand.Shuffle(len(player2Deck), func(i, j int) {
		player2Deck[i], player2Deck[j] = player2Deck[j], player2Deck[i]
	})

	// プレイヤーを作成
	player1 := entity.NewPlayer(player1ID, player1Name, player1Deck)
	player2 := entity.NewPlayer(player2ID, player2Name, player2Deck)

	// 初期手札をドロー
	if _, err := player1.DrawCards(entity.InitialHandSize); err != nil {
		return fmt.Errorf("failed to draw initial hand for player1: %w", err)
	}
	if _, err := player2.DrawCards(entity.InitialHandSize); err != nil {
		return fmt.Errorf("failed to draw initial hand for player2: %w", err)
	}

	// ゲーム状態を作成
	state := game.NewState(gameID, player1, player2, s.logger)

	// ゲームセッションを作成
	session := &GameSession{
		State:            state,
		UsecaseEngine:    usecase.NewEngine(state),
		EffectProcessor:  effect.NewProcessor(state),
		ConnectedPlayers: make(map[string]bool),
	}

	s.games[gameID] = session

	// 初期化イベントを送信
	s.eventBroadcaster.Broadcast(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "game_created",
		Message:   "Game created successfully",
		State:     state,
	})

	return nil
}

// GetGameState は指定されたゲームの現在の状態を取得
func (s *GameService) GetGameState(gameID string) (*game.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.games[gameID]
	if !exists {
		return nil, entity.NewErrNotFound("game", gameID)
	}

	return session.State, nil
}

// PlayCard は指定されたプレイヤーがカードをプレイ
func (s *GameService) PlayCard(ctx context.Context, gameID string, playerID string, cardID string, targetID *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.games[gameID]
	if !exists {
		return entity.NewErrNotFound("game", gameID)
	}

	state := session.State
	player := state.GetPlayerByID(playerID)
	if player == nil {
		return entity.NewErrNotFound("player", playerID)
	}

	// アクションの妥当性を検証
	if err := state.ValidateAction(playerID); err != nil {
		return err
	}

	// カードを手札から検索
	var targetCard *entity.Card
	for _, card := range player.Hand {
		if card.ID == cardID {
			targetCard = &card
			break
		}
	}

	if targetCard == nil {
		return entity.NewErrNotFound("card", cardID)
	}

	s.logger.Info("PlayCard: %s (Type=%s, IsUnit=%v)", targetCard.Name, targetCard.Type, targetCard.IsUnit())

	// デバッグ: スペルの場合、CardEffectとtargetIDを確認
	if targetCard.IsSpell() {
		s.logger.Info("Spell Debug - CardEffect: %v, targetID: %v", targetCard.CardEffect != nil, targetID)
		if targetCard.CardEffect != nil && len(targetCard.CardEffect.Definitions) > 0 {
			s.logger.Info("Spell Debug - RequireTarget: %v", targetCard.CardEffect.Definitions[0].RequireTarget)
		}
	}

	// ユースケース層にビジネスロジックを委譲
	if targetCard.IsUnit() {
		card, err := session.UsecaseEngine.SummonUnit(playerID, cardID)
		if err != nil {
			return err
		}

		// ログ追加
		state.AddLog(playerID, "ユニット召喚", fmt.Sprintf("%s を召喚", card.Name))

		// 召喚時効果を処理
		if err := session.EffectProcessor.ProcessTimingEffects(card, entity.EffectTimingOnSummon, playerID, nil); err != nil {
			// エラーをログに記録して継続（効果失敗でもゲームは続行）
			s.logger.Error("召喚時効果エラー: %v", err)
			state.AddLog(playerID, "エラー", fmt.Sprintf("召喚時効果の処理に失敗: %v", err))
		}
	} else {
		card, err := session.UsecaseEngine.UseSpell(playerID, cardID, targetID)
		if err != nil {
			return err
		}

		s.logger.Info("「%s」を使用 (コスト: %d) [残りマナ: %d]", card.Name, card.Cost, player.CurrentTurnMana)
		s.logger.Info("効果: %s", card.Effect)

		state.AddLog(playerID, "スペル使用", fmt.Sprintf("%s を使用", card.Name))

		if card.CardEffect != nil {
			for _, def := range card.CardEffect.Definitions {
				if err := session.EffectProcessor.ProcessEffectDefinition(def, playerID, targetID); err != nil {
					s.logger.Error("効果処理エラー: %v", err)
					return err
				}
			}
		}
	}

	// アクションイベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "card_played",
		Message:   fmt.Sprintf("%s played %s", playerID, targetCard.Name),
		PlayerID:  playerID,
		Timestamp: time.Now(),
		State:     state,
	})

	return nil
}

// ExecuteAttack は指定されたユニットで攻撃
func (s *GameService) ExecuteAttack(ctx context.Context, gameID string, attackerPlayerID string, attackerID string, targetID *string) (*entity.CombatResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game %s not found", gameID)
	}

	state := session.State

	// アクションの妥当性を検証
	if err := state.ValidateAction(attackerPlayerID); err != nil {
		return nil, err
	}

	// AttackActionを構築してユースケース層に委譲
	action := entity.AttackAction{
		PlayerID:   attackerPlayerID,
		AttackerID: attackerID,
		TargetID:   targetID,
	}

	// 攻撃を実行
	result, err := session.UsecaseEngine.ExecuteAttack(action)
	if err != nil {
		return nil, err
	}

	// 攻撃イベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "attack_executed",
		Message:   fmt.Sprintf("%s executed attack", attackerPlayerID),
		PlayerID:  attackerPlayerID,
		Timestamp: time.Now(),
		State:     state,
	})

	return result, nil
}

// EndTurn はターンを終了
func (s *GameService) EndTurn(ctx context.Context, gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.games[gameID]
	if !exists {
		return entity.NewErrNotFound("game", gameID)
	}

	// 現在のプレイヤーを取得してターン終了処理を委譲
	state := session.State
	currentPlayer := state.GetCurrentPlayer()
	if currentPlayer == nil {
		return entity.NewErrNotFound("player", "current_player")
	}

	if err := session.UsecaseEngine.EndTurn(currentPlayer.ID); err != nil {
		return err
	}

	// ゲーム終了チェックとイベント送信
	if state.IsOver() {
		s.broadcastEvent(gameID, &event.GameEvent{
			GameID:    gameID,
			EventType: "game_over",
			Message:   "Game Over",
			PlayerID:  state.CurrentPlayerID,
			Timestamp: time.Now(),
			State:     state,
		})
		return nil
	}

	// フェーズ変更イベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "turn_ended",
		Message:   fmt.Sprintf("Turn ended, now %s's turn", state.CurrentPlayerID),
		PlayerID:  state.CurrentPlayerID,
		Timestamp: time.Now(),
		State:     state,
	})

	return nil
}

// SubscribeToEvents はゲームイベントのストリームを購読
func (s *GameService) SubscribeToEvents(gameID string) (chan *event.GameEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.games[gameID]; !exists {
		return nil, fmt.Errorf("game %s not found", gameID)
	}

	return s.eventBroadcaster.Subscribe(gameID), nil
}

// UnsubscribeFromEvents はイベントストリームの購読を解除
func (s *GameService) UnsubscribeFromEvents(gameID string, eventChan chan *event.GameEvent) {
	s.eventBroadcaster.Unsubscribe(gameID, eventChan)
}

// MarkPlayerConnected プレイヤーを接続状態にマーク
func (s *GameService) MarkPlayerConnected(gameID, playerID string) error {
	s.mu.RLock()
	session, exists := s.games[gameID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("game %s not found", gameID)
	}

	session.mu.Lock()
	session.ConnectedPlayers[playerID] = true
	session.mu.Unlock()

	s.logger.Info("🔌 Player %s connected to game %s", playerID, gameID)
	return nil
}

// MarkPlayerDisconnected プレイヤーを切断状態にマークし、相手に通知
func (s *GameService) MarkPlayerDisconnected(gameID, playerID string) error {
	s.mu.RLock()
	session, exists := s.games[gameID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("game %s not found", gameID)
	}

	session.mu.Lock()
	delete(session.ConnectedPlayers, playerID)
	session.mu.Unlock()

	s.logger.Info("🔌 Player %s disconnected from game %s", playerID, gameID)

	// 切断イベントをブロードキャスト
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "player_disconnected",
		Message:   fmt.Sprintf("Player %s has disconnected", playerID),
		PlayerID:  playerID,
		Timestamp: time.Now(),
		State:     session.State,
	})

	return nil
}

// DeleteGame はゲームセッションを削除
func (s *GameService) DeleteGame(gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[gameID]; !exists {
		return fmt.Errorf("game %s not found", gameID)
	}

	// イベントブロードキャスターのクリーンアップ
	s.eventBroadcaster.CleanupGame(gameID)

	delete(s.games, gameID)

	return nil
}

// broadcastEvent はゲームイベントをすべての購読者に送信
func (s *GameService) broadcastEvent(gameID string, evt *event.GameEvent) {
	s.eventBroadcaster.Broadcast(gameID, evt)
}

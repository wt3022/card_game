package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/core/usecase"
	"card_game/internal/core/usecase/effect"
	"card_game/internal/core/usecase/game"
	fixturedeck "card_game/internal/fixture/deck"
	"card_game/internal/infrastructure/event"
)

// GameSession はゲームセッションを表す構造体
type GameSession struct {
	State           *game.State
	UsecaseEngine   *usecase.Engine
	EffectProcessor *effect.Processor
}

// GameService はゲームセッションを管理するサービス
type GameService struct {
	mu               sync.RWMutex
	games            map[string]*GameSession
	eventBroadcaster *event.Broadcaster
	logger           port.Logger
	waitingPlayers   []*MatchmakingPlayer
}

func (s *GameService) MarkPlayerDisconnected(gameID string, playerID string) {
	// 現状は何もしない（将来の拡張用）
}

func (s *GameService) MarkPlayerConnected(gameID string, playerID string) error {
	// 現状は何もしない（将来の拡張用）
	return nil
}

// MatchmakingPlayer マッチング待機中のプレイヤー
type MatchmakingPlayer struct {
	PlayerID   string
	PlayerName string
	JoinedAt   time.Time
	NotifyChan chan *MatchResult
}

// MatchResult マッチング結果
type MatchResult struct {
	GameID    string
	Player1ID string
	Player2ID string
	Success   bool
	Message   string
}

// NewGameService は新しいGameServiceを作成
func NewGameService(logger port.Logger) *GameService {
	return &GameService{
		games:            make(map[string]*GameSession),
		eventBroadcaster: event.NewBroadcaster(),
		logger:           logger,
		waitingPlayers:   []*MatchmakingPlayer{},
	}
}

// CreateGame は新しいゲームセッションを作成
func (s *GameService) CreateGame(gameID string, player1ID, player1Name, player2ID, player2Name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[gameID]; exists {
		return entity.NewErrAlreadyExists("game", gameID)
	}

	// デッキを生成
	player1Deck := fixturedeck.GenerateSampleDeck(player1ID)
	player2Deck := fixturedeck.GenerateSampleDeck(player2ID)

	// デッキをシャッフル
	player1Deck = game.ShuffleDeck(player1Deck)
	player2Deck = game.ShuffleDeck(player2Deck)

	// プレイヤーを作成
	player1 := entity.NewPlayer(player1ID, player1Name, player1Deck)
	player2 := entity.NewPlayer(player2ID, player2Name, player2Deck)
	// 先攻のみIsFirstTurn=true、後攻はfalse
	player1.IsFirstTurn = true
	player2.IsFirstTurn = false

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
		State:           state,
		UsecaseEngine:   usecase.NewEngine(state),
		EffectProcessor: effect.NewProcessor(state),
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

// PerformMulligan 指定されたプレイヤーのマリガンを実行
func (s *GameService) PerformMulligan(ctx context.Context, gameID string, playerID string, cardIDs []string) error {
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

	// すでにマリガン済みかチェック
	if state.IsMulliganDone(playerID) {
		return entity.NewErrInvalidState("mulligan", "マリガンは既に完了しています")
	}

	// マリガンを実行
	drawnCards, err := player.PerformMulligan(cardIDs, game.ShuffleDeck)
	if err != nil {
		return err
	}

	// マリガン完了状態を設定
	if err := state.SetMulliganDone(playerID); err != nil {
		return err
	}

	// ログを追加
	if len(cardIDs) > 0 {
		state.AddLog(playerID, "マリガン", fmt.Sprintf("%d枚のカードを交換しました", len(cardIDs)))
		s.logger.Info("%s: %d枚のカードをマリガン", player.Name, len(cardIDs))
	} else {
		state.AddLog(playerID, "マリガン", "マリガンをスキップしました")
		s.logger.Info("%s: マリガンをスキップ", player.Name)
	}

	// 引いたカードのログ
	for _, card := range drawnCards {
		s.logger.Info("%s: マリガンで「%s」を引きました", player.Name, card.Name)
	}

	// マリガンイベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "mulligan_performed",
		Message:   fmt.Sprintf("%s performed mulligan", playerID),
		PlayerID:  playerID,
		Timestamp: time.Now(),
		State:     state,
	})

	// 両プレイヤーのマリガンが完了したら、ゲームを開始
	if state.AreBothPlayersMulliganDone() {
		s.logger.Info("両プレイヤーのマリガンが完了しました。ゲームを開始します。")

		// マリガン完了イベントを送信
		s.broadcastEvent(gameID, &event.GameEvent{
			GameID:    gameID,
			EventType: "mulligan_completed",
			Message:   "両プレイヤーのマリガンが完了しました",
			Timestamp: time.Now(),
			State:     state,
		})

		// 先攻プレイヤーのターンを自動的に開始
		currentPlayer := state.GetCurrentPlayer()
		if currentPlayer != nil {
			if err := session.UsecaseEngine.StartTurn(currentPlayer.ID); err != nil {
				s.logger.Error("最初のターン開始エラー: %v", err)
				return err
			}

			// ターン開始イベントを送信
			s.broadcastEvent(gameID, &event.GameEvent{
				GameID:    gameID,
				EventType: "turn_started",
				Message:   fmt.Sprintf("Turn started for %s", currentPlayer.ID),
				PlayerID:  currentPlayer.ID,
				Timestamp: time.Now(),
				State:     state,
			})

			s.logger.Info("ゲーム開始: %sのターン", currentPlayer.Name)
		}
	}

	return nil
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

// StartTurn はターンを開始
func (s *GameService) StartTurn(ctx context.Context, gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.games[gameID]
	if !exists {
		return entity.NewErrNotFound("game", gameID)
	}

	state := session.State
	currentPlayer := state.GetCurrentPlayer()
	if currentPlayer == nil {
		return entity.NewErrNotFound("player", "current_player")
	}
	if err := session.UsecaseEngine.StartTurn(currentPlayer.ID); err != nil {
		return err
	}

	// ターン開始イベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "turn_started",
		Message:   fmt.Sprintf("Turn started for %s", currentPlayer.ID),
		PlayerID:  currentPlayer.ID,
		Timestamp: time.Now(),
		State:     state,
	})
	return nil
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

	// ターン終了イベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "turn_ended",
		Message:   fmt.Sprintf("Turn ended, now %s's turn", state.CurrentPlayerID),
		PlayerID:  state.CurrentPlayerID,
		Timestamp: time.Now(),
		State:     state,
	})

	// 次のプレイヤーのターンを自動的に開始
	nextPlayer := state.GetCurrentPlayer()
	if nextPlayer == nil {
		return entity.NewErrNotFound("player", "next_player")
	}

	if err := session.UsecaseEngine.StartTurn(nextPlayer.ID); err != nil {
		return err
	}

	// ターン開始イベントを送信
	s.broadcastEvent(gameID, &event.GameEvent{
		GameID:    gameID,
		EventType: "turn_started",
		Message:   fmt.Sprintf("Turn started for %s", nextPlayer.ID),
		PlayerID:  nextPlayer.ID,
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

// ========================================
// マッチメイキング機能
// ========================================

// JoinQueue プレイヤーをマッチングキューに追加
func (s *GameService) JoinQueue(playerID, playerName string) chan *MatchResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 通知用チャネルを作成
	notifyChan := make(chan *MatchResult, 1)

	player := &MatchmakingPlayer{
		PlayerID:   playerID,
		PlayerName: playerName,
		JoinedAt:   time.Now(),
		NotifyChan: notifyChan,
	}

	s.logger.Info("🎮 プレイヤーがマッチングキューに参加: %s (%s)", playerName, playerID)

	// すでに待機中のプレイヤーがいればマッチング
	if len(s.waitingPlayers) > 0 {
		opponent := s.waitingPlayers[0]
		s.waitingPlayers = s.waitingPlayers[1:]

		s.logger.Info("✨ マッチング成功: %s vs %s", opponent.PlayerName, playerName)

		// ゲームを作成
		go s.createMatchedGame(opponent, player)
	} else {
		// 待機リストに追加
		s.waitingPlayers = append(s.waitingPlayers, player)
		s.logger.Info("⏳ 待機中のプレイヤー数: %d", len(s.waitingPlayers))
	}

	return notifyChan
}

// createMatchedGame マッチング成功後にゲームを作成
func (s *GameService) createMatchedGame(player1, player2 *MatchmakingPlayer) {
	// ゲームIDを生成
	gameID := fmt.Sprintf("game-%s-%s", player1.PlayerID, player2.PlayerID)

	// ゲームを作成
	err := s.CreateGame(
		gameID,
		player1.PlayerID,
		player1.PlayerName,
		player2.PlayerID,
		player2.PlayerName,
	)

	if err != nil {
		s.logger.Error("ゲーム作成エラー: %v", err)

		// 両プレイヤーにエラーを通知
		result := &MatchResult{
			Success: false,
			Message: fmt.Sprintf("ゲーム作成に失敗しました: %v", err),
		}
		player1.NotifyChan <- result
		player2.NotifyChan <- result
		return
	}

	// 両プレイヤーに成功を通知
	result1 := &MatchResult{
		GameID:    gameID,
		Player1ID: player1.PlayerID,
		Player2ID: player2.PlayerID,
		Success:   true,
		Message:   "マッチング成功！",
	}
	result2 := &MatchResult{
		GameID:    gameID,
		Player1ID: player1.PlayerID,
		Player2ID: player2.PlayerID,
		Success:   true,
		Message:   "マッチング成功！",
	}

	player1.NotifyChan <- result1
	player2.NotifyChan <- result2

	s.logger.Info("🎉 ゲーム作成完了: %s", gameID)
}

// LeaveQueue プレイヤーをマッチングキューから削除
func (s *GameService) LeaveQueue(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// プレイヤーを待機リストから削除
	for i, player := range s.waitingPlayers {
		if player.PlayerID == playerID {
			s.waitingPlayers = append(s.waitingPlayers[:i], s.waitingPlayers[i+1:]...)
			s.logger.Info("👋 プレイヤーがマッチングキューから退出: %s (%s)", player.PlayerName, playerID)
			close(player.NotifyChan) // チャネルをクローズ
			return
		}
	}
}

// GetWaitingCount 待機中のプレイヤー数を取得
func (s *GameService) GetWaitingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.waitingPlayers)
}

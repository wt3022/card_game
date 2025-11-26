package handler

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/converter"
	"card_game/internal/application/service"
	"card_game/internal/core/entity"
	"card_game/internal/infrastructure/middleware"
	"card_game/internal/infrastructure/security"
)

// ========================================
// Connect-Goハンドラー
// protoで定義されたGameServiceを実装
// 設計方針:
// - Connect-Goの標準的なハンドラーパターン
// - application/serviceを利用
// - protoとdomainエンティティのマッピング
// ========================================

// GameConnectHandler Connect-Go用のゲームサービスハンドラー
type GameConnectHandler struct {
	gameService    *service.GameService
	rateLimiter    *middleware.RateLimiter
	stateValidator *security.StateValidator
	cheatDetector  *security.CheatDetector
	lastActionTime map[string]time.Time
}

// NewGameConnectHandler 新しいGameConnectHandlerを作成
func NewGameConnectHandler(gameService *service.GameService) *GameConnectHandler {
	rateLimiter := middleware.NewRateLimiter()
	rateLimiter.StartCleanupRoutine()

	return &GameConnectHandler{
		gameService:    gameService,
		rateLimiter:    rateLimiter,
		stateValidator: security.NewStateValidator(),
		cheatDetector:  security.NewCheatDetector(),
		lastActionTime: make(map[string]time.Time),
	}
}

// インターフェースの実装確認
var _ cardgamev1connect.GameServiceHandler = (*GameConnectHandler)(nil)

// CreateGame ゲームを新規作成
func (h *GameConnectHandler) CreateGame(
	ctx context.Context,
	req *connect.Request[pbv1.CreateGameRequest],
) (*connect.Response[pbv1.CreateGameResponse], error) {
	// リクエストからパラメータを取得
	player1ID := req.Msg.GetPlayer1Id()
	player2ID := req.Msg.GetPlayer2Id()
	player1Name := req.Msg.GetPlayer1Name()
	player2Name := req.Msg.GetPlayer2Name()

	// ゲームIDを生成（通常はUUID等を使用）
	gameID := fmt.Sprintf("game-%s-%s", player1ID, player2ID)

	// ゲームを作成
	err := h.gameService.CreateGame(gameID, player1ID, player1Name, "", player2ID, player2Name, "")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// ゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスを作成
	resp := &pbv1.CreateGameResponse{
		GameId:    gameID,
		GameState: converter.GameStateToProto(state, player1ID),
	}

	return connect.NewResponse(resp), nil
}

// GetGameState ゲームの現在の状態を取得
func (h *GameConnectHandler) GetGameState(
	ctx context.Context,
	req *connect.Request[pbv1.GetGameStateRequest],
) (*connect.Response[pbv1.GetGameStateResponse], error) {
	gameID := req.Msg.GetGameId()

	// ゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	playerID := req.Msg.GetPlayerId()

	// レスポンスを作成
	resp := &pbv1.GetGameStateResponse{
		GameState: converter.GameStateToProto(state, playerID),
	}

	return connect.NewResponse(resp), nil
}

// PerformMulligan マリガンを実行
func (h *GameConnectHandler) PerformMulligan(
	ctx context.Context,
	req *connect.Request[pbv1.PerformMulliganRequest],
) (*connect.Response[pbv1.PerformMulliganResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()
	cardIDs := req.Msg.GetCardIds()

	// レート制限チェック（マリガンは厳しめ）
	if err := h.rateLimiter.CheckLimit(playerID, "PerformMulligan"); err != nil {
		return nil, connect.NewError(connect.CodeResourceExhausted, err)
	}

	// マリガンを実行
	err := h.gameService.PerformMulligan(ctx, gameID, playerID, cardIDs)
	if err != nil {
		return nil, mapDomainErrorToConnectError(err)
	}

	// 更新されたゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスを作成
	resp := &pbv1.PerformMulliganResponse{
		Success:   true,
		Message:   "Mulligan performed successfully",
		GameState: converter.GameStateToProto(state, playerID),
	}

	return connect.NewResponse(resp), nil
}

// PlayCard カードをプレイ
func (h *GameConnectHandler) PlayCard(
	ctx context.Context,
	req *connect.Request[pbv1.PlayCardRequest],
) (*connect.Response[pbv1.PlayCardResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()
	cardID := req.Msg.GetCardId()

	// レート制限チェック
	if err := h.rateLimiter.CheckLimit(playerID, "PlayCard"); err != nil {
		return nil, connect.NewError(connect.CodeResourceExhausted, err)
	}

	// 高速操作の検出
	if lastTime, ok := h.lastActionTime[playerID]; ok {
		h.cheatDetector.DetectImpossibleTiming(playerID, "PlayCard", lastTime)
	}
	h.lastActionTime[playerID] = time.Now()

	var targetID *string
	if req.Msg.TargetId != nil {
		targetID = req.Msg.TargetId
	}

	// カードをプレイ
	err := h.gameService.PlayCard(ctx, gameID, playerID, cardID, targetID)
	if err != nil {
		// ドメインエラーをConnect-Goエラーに変換
		return nil, mapDomainErrorToConnectError(err)
	}

	// 更新されたゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスを作成
	resp := &pbv1.PlayCardResponse{
		Success:   true,
		Message:   "Card played successfully",
		GameState: converter.GameStateToProto(state, playerID),
	}

	return connect.NewResponse(resp), nil
}

// ExecuteAttack 攻撃を実行
func (h *GameConnectHandler) ExecuteAttack(
	ctx context.Context,
	req *connect.Request[pbv1.ExecuteAttackRequest],
) (*connect.Response[pbv1.ExecuteAttackResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()
	attackerID := req.Msg.GetAttackerId()

	// レート制限チェック
	if err := h.rateLimiter.CheckLimit(playerID, "ExecuteAttack"); err != nil {
		return nil, connect.NewError(connect.CodeResourceExhausted, err)
	}

	// 高速操作の検出
	if lastTime, ok := h.lastActionTime[playerID]; ok {
		h.cheatDetector.DetectImpossibleTiming(playerID, "ExecuteAttack", lastTime)
	}
	h.lastActionTime[playerID] = time.Now()

	var targetID *string
	if req.Msg.TargetId != nil {
		targetID = req.Msg.TargetId
	}

	// 攻撃を実行
	result, err := h.gameService.ExecuteAttack(ctx, gameID, playerID, attackerID, targetID)
	if err != nil {
		return nil, mapDomainErrorToConnectError(err)
	}

	// 更新されたゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスを作成
	resp := &pbv1.ExecuteAttackResponse{
		Success:   true,
		Message:   "Attack executed successfully",
		Result:    converter.CombatResultToProto(result),
		GameState: converter.GameStateToProto(state, playerID),
	}

	return connect.NewResponse(resp), nil
}

// StartTurn ターンを開始
func (h *GameConnectHandler) StartTurn(
	ctx context.Context,
	req *connect.Request[pbv1.StartTurnRequest],
) (*connect.Response[pbv1.StartTurnResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()

	// ターンを開始
	err := h.gameService.StartTurn(ctx, gameID)
	if err != nil {
		return nil, mapDomainErrorToConnectError(err)
	}

	// 更新されたゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスを作成
	resp := &pbv1.StartTurnResponse{
		Success:   true,
		Message:   "Turn started successfully",
		GameState: converter.GameStateToProto(state, playerID),
	}

	return connect.NewResponse(resp), nil
}

// EndTurn ターンを終了
func (h *GameConnectHandler) EndTurn(
	ctx context.Context,
	req *connect.Request[pbv1.EndTurnRequest],
) (*connect.Response[pbv1.EndTurnResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()

	// レート制限チェック
	if err := h.rateLimiter.CheckLimit(playerID, "EndTurn"); err != nil {
		return nil, connect.NewError(connect.CodeResourceExhausted, err)
	}

	// ターンを終了
	err := h.gameService.EndTurn(ctx, gameID)
	if err != nil {
		return nil, mapDomainErrorToConnectError(err)
	}

	// 更新されたゲーム状態を取得
	state, err := h.gameService.GetGameState(gameID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスを作成
	resp := &pbv1.EndTurnResponse{
		Success:   true,
		Message:   "Turn ended successfully",
		GameState: converter.GameStateToProto(state, playerID),
	}

	return connect.NewResponse(resp), nil
}

// StreamGameEvents ゲームイベントをストリーミング（サーバーサイド）
func (h *GameConnectHandler) StreamGameEvents(
	ctx context.Context,
	req *connect.Request[pbv1.GameEventRequest],
	stream *connect.ServerStream[pbv1.GameEventResponse],
) error {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()

	fmt.Printf("🔌 ゲーム %s のプレイヤー %s がイベントストリームに接続しました\n", gameID, playerID)

	// イベントをSubscribe
	eventChan, err := h.gameService.SubscribeToEvents(gameID)
	if err != nil {
		return err
	}
	defer h.gameService.UnsubscribeFromEvents(gameID, eventChan)

	// プレイヤーを接続状態にマーク
	if err := h.gameService.MarkPlayerConnected(gameID, playerID); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	// 接続後の最新ゲーム状態を取得して送信（接続状態の更新を反映）
	state, err := h.gameService.GetGameState(gameID)
	if err == nil {
		initialResp := &pbv1.GameEventResponse{
			Event:     nil, // 初期状態送信なのでイベントはnil
			GameState: converter.GameStateToProto(state, playerID),
		}
		if err := stream.Send(initialResp); err != nil {
			return err
		}
	}

	// 切断時に切断状態をマーク
	defer func() {
		fmt.Printf("🔌 プレイヤー %s (ゲーム %s) のイベントストリームを終了します\n", playerID, gameID)
		h.gameService.MarkPlayerDisconnected(gameID, playerID)
	}()

	for {
		select {
		case <-ctx.Done():
			// 接続が切れた場合（defer で MarkPlayerDisconnected が呼ばれる）
			fmt.Printf("⚠️ プレイヤー %s (ゲーム %s) のコンテキストが終了しました: %v\n", playerID, gameID, ctx.Err())
			return ctx.Err()
		case event := <-eventChan:
			// イベントをprotoに変換して送信
			resp := &pbv1.GameEventResponse{
				Event:     converter.GameEventToProto(event),
				GameState: converter.GameStateToProto(event.State, playerID),
			}
			if err := stream.Send(resp); err != nil {
				// 送信エラー時も切断とみなす（defer で MarkPlayerDisconnected が呼ばれる）
				return err
			}
		}
	}
}

// ========================================
// ヘルパー関数
// ========================================

// mapDomainErrorToConnectError ドメインエラーをConnect-Goエラーに変換
func mapDomainErrorToConnectError(err error) error {
	domainErr, ok := entity.AsDomainError(err)
	if !ok {
		return connect.NewError(connect.CodeInternal, err)
	}

	switch domainErr.Category() {
	case entity.ErrorCategoryNotFound:
		return connect.NewError(connect.CodeNotFound, err)
	case entity.ErrorCategoryInvalidInput:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case entity.ErrorCategoryPrecondition:
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case entity.ErrorCategoryConflict:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case entity.ErrorCategoryInternal:
		return connect.NewError(connect.CodeInternal, err)
	default:
		return connect.NewError(connect.CodeUnknown, err)
	}
}

// JoinMatchmaking マッチメイキングに参加（サーバーストリーミング）
func (h *GameConnectHandler) JoinMatchmaking(
	ctx context.Context,
	req *connect.Request[pbv1.JoinMatchmakingRequest],
	stream *connect.ServerStream[pbv1.MatchmakingResponse],
) error {
	playerID := req.Msg.GetPlayerId()
	playerName := req.Msg.GetPlayerName()
	deckID := ""
	if req.Msg.DeckId != nil {
		deckID = *req.Msg.DeckId
	}

	// マッチングキューに参加
	notifyChan := h.gameService.JoinQueue(playerID, playerName, deckID)

	// 接続が切れた時にキューから退出させる
	defer h.gameService.LeaveQueue(playerID)

	// 待機中のステータスを送信
	err := stream.Send(&pbv1.MatchmakingResponse{
		Status:  pbv1.MatchmakingStatus_MATCHMAKING_STATUS_WAITING,
		Message: "マッチング相手を探しています...",
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	// マッチング結果を待機
	select {
	case result := <-notifyChan:
		if !result.Success {
			return stream.Send(&pbv1.MatchmakingResponse{
				Status:  pbv1.MatchmakingStatus_MATCHMAKING_STATUS_UNSPECIFIED,
				Message: result.Message,
			})
		}

		// マッチング成功を通知
		err = stream.Send(&pbv1.MatchmakingResponse{
			Status:  pbv1.MatchmakingStatus_MATCHMAKING_STATUS_MATCHED,
			GameId:  result.GameID,
			Message: result.Message,
		})
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}

		// ゲーム状態を取得して送信
		gameState, err := h.gameService.GetGameState(result.GameID)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}

		protoState := converter.GameStateToProto(gameState, playerID)
		return stream.Send(&pbv1.MatchmakingResponse{
			Status:    pbv1.MatchmakingStatus_MATCHMAKING_STATUS_GAME_STARTED,
			GameId:    result.GameID,
			GameState: protoState,
			Message:   "ゲーム開始！",
		})

	case <-ctx.Done():
		// 接続が切れた場合は Canceled エラーを返す（defer で LeaveQueue が呼ばれる）
		return connect.NewError(connect.CodeCanceled, ctx.Err())
	}
}

func ShuffleDeck(deck []entity.Card) []entity.Card {
	shuffled_deck := make([]entity.Card, len(deck))

	round := rand.IntN(len(deck))
	// ランダムな回数だけシャッフルを繰り返す
	for range round {
		perm := rand.Perm(len(deck))
		for i, v := range perm {
			shuffled_deck[i] = deck[v]
		}
		deck = make([]entity.Card, len(shuffled_deck))
		copy(deck, shuffled_deck)
	}
	return shuffled_deck
}

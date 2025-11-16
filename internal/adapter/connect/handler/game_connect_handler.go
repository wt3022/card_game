package handler

import (
	"context"
	"fmt"
	"math/rand/v2"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/converter"
	"card_game/internal/application/service"
	"card_game/internal/core/entity"
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
	gameService        *service.GameService
	matchmakingService *service.MatchmakingService
}

// NewGameConnectHandler 新しいGameConnectHandlerを作成
func NewGameConnectHandler(gameService *service.GameService, matchmakingService *service.MatchmakingService) *GameConnectHandler {
	return &GameConnectHandler{
		gameService:        gameService,
		matchmakingService: matchmakingService,
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

	// サンプルデッキを生成
	player1Deck := generateSampleDeck(player1ID)
	player2Deck := generateSampleDeck(player2ID)

	// ゲームを作成
	err := h.gameService.CreateGame(gameID, player1ID, player1Name, player2ID, player2Name, player1Deck, player2Deck)
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

// PlayCard カードをプレイ
func (h *GameConnectHandler) PlayCard(
	ctx context.Context,
	req *connect.Request[pbv1.PlayCardRequest],
) (*connect.Response[pbv1.PlayCardResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()
	cardID := req.Msg.GetCardId()

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

// EndTurn ターンを終了
func (h *GameConnectHandler) EndTurn(
	ctx context.Context,
	req *connect.Request[pbv1.EndTurnRequest],
) (*connect.Response[pbv1.EndTurnResponse], error) {
	gameID := req.Msg.GetGameId()
	playerID := req.Msg.GetPlayerId()

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

	// 切断時に切断状態をマーク
	defer h.gameService.MarkPlayerDisconnected(gameID, playerID)

	for {
		select {
		case <-ctx.Done():
			// 接続が切れた場合（defer で MarkPlayerDisconnected が呼ばれる）
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

// createSpellCard スペルカードを作成（説明文を自動生成）
func createSpellCard(id, name string, cost int, effect *entity.CardEffect) entity.Card {
	// 説明文を自動生成
	for _, def := range effect.Definitions {
		def.Description = def.GenerateDescription()
	}
	effect.Description = effect.GenerateDescription()

	return entity.Card{
		ID:         id,
		Name:       name,
		Cost:       cost,
		Type:       entity.CardTypeSpell,
		CardEffect: effect,
	}
}

// GenerateSampleDeckForTest テスト用にエクスポートされたデッキ生成関数
func GenerateSampleDeckForTest(prefix string) []entity.Card {
	return generateSampleDeck(prefix)
}

// generateSampleDeck サンプルデッキを生成（40枚、多様な要素を含む）
func generateSampleDeck(prefix string) []entity.Card {
	deck := []entity.Card{}

	// ========================================
	// 通常ユニットカード（12枚）
	// コスト1-10で山形分布
	// ========================================

	// 低コスト (1-2コスト): 3枚
	attack1, defense1 := 2, 1
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-goblin", prefix),
		Name:    "Goblin Scout",
		Cost:    1,
		Type:    entity.CardTypeUnit,
		Attack:  &attack1,
		Defense: &defense1,
		Traits:  []entity.Trait{},
	})

	attack2, defense2 := 2, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-soldier", prefix),
		Name:    "Soldier",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attack2,
		Defense: &defense2,
		Traits:  []entity.Trait{},
	})

	attack3, defense3 := 3, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-archer", prefix),
		Name:    "Archer",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attack3,
		Defense: &defense3,
		Traits:  []entity.Trait{},
	})

	// 中コスト (3-5コスト): 6枚
	attack4, defense4 := 3, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-warrior", prefix),
		Name:    "Warrior",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attack4,
		Defense: &defense4,
		Traits:  []entity.Trait{},
	})

	attack5, defense5 := 4, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-knight", prefix),
		Name:    "Knight",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attack5,
		Defense: &defense5,
		Traits:  []entity.Trait{},
	})

	attack6, defense6 := 3, 5
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-defender", prefix),
		Name:    "Defender",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attack6,
		Defense: &defense6,
		Traits:  []entity.Trait{},
	})

	attack7, defense7 := 5, 4
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-champion", prefix),
		Name:    "Champion",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attack7,
		Defense: &defense7,
		Traits:  []entity.Trait{},
	})

	attack8, defense8 := 4, 5
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-paladin", prefix),
		Name:    "Paladin",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attack8,
		Defense: &defense8,
		Traits:  []entity.Trait{},
	})

	attack9, defense9 := 5, 5
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-general", prefix),
		Name:    "General",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attack9,
		Defense: &defense9,
		Traits:  []entity.Trait{},
	})

	// 高コスト (6-10コスト): 3枚
	attack10, defense10 := 6, 6
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-warlord", prefix),
		Name:    "Warlord",
		Cost:    7,
		Type:    entity.CardTypeUnit,
		Attack:  &attack10,
		Defense: &defense10,
		Traits:  []entity.Trait{},
	})

	attack11, defense11 := 7, 7
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-colossus", prefix),
		Name:    "Colossus",
		Cost:    8,
		Type:    entity.CardTypeUnit,
		Attack:  &attack11,
		Defense: &defense11,
		Traits:  []entity.Trait{},
	})

	attack12, defense12 := 10, 10
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-titan", prefix),
		Name:    "Ancient Titan",
		Cost:    10,
		Type:    entity.CardTypeUnit,
		Attack:  &attack12,
		Defense: &defense12,
		Traits:  []entity.Trait{},
	})

	// ========================================
	// 特殊能力持ちユニット（12枚）
	// ========================================

	// Rush (疾走) - 2枚
	attackRush1, defenseRush1 := 3, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-rush-knight", prefix),
		Name:    "Rush Knight",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush1,
		Defense: &defenseRush1,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	attackRush2, defenseRush2 := 4, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-rush-cavalry", prefix),
		Name:    "Swift Cavalry",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush2,
		Defense: &defenseRush2,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	// Guardian (守護) - 2枚
	attackGuard1, defenseGuard1 := 2, 5
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-guardian-shield", prefix),
		Name:    "Shield Guardian",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard1,
		Defense: &defenseGuard1,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	attackGuard2, defenseGuard2 := 3, 6
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-guardian-wall", prefix),
		Name:    "Stone Wall",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard2,
		Defense: &defenseGuard2,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	// Windfury (疾風) - 2枚
	attackWind1, defenseWind1 := 2, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-wind-striker", prefix),
		Name:    "Wind Striker",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackWind1,
		Defense: &defenseWind1,
		Traits:  []entity.Trait{entity.TraitWindfury},
	})

	attackWind2, defenseWind2 := 3, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-wind-master", prefix),
		Name:    "Wind Master",
		Cost:    6,
		Type:    entity.CardTypeUnit,
		Attack:  &attackWind2,
		Defense: &defenseWind2,
		Traits:  []entity.Trait{entity.TraitWindfury},
	})

	// Pierce (貫通) - 2枚
	attackPierce1, defensePierce1 := 4, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-pierce-lancer", prefix),
		Name:    "Lance Piercer",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackPierce1,
		Defense: &defensePierce1,
		Traits:  []entity.Trait{entity.TraitPierce},
	})

	attackPierce2, defensePierce2 := 5, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-pierce-dragon", prefix),
		Name:    "Pierce Dragon",
		Cost:    6,
		Type:    entity.CardTypeUnit,
		Attack:  &attackPierce2,
		Defense: &defensePierce2,
		Traits:  []entity.Trait{entity.TraitPierce},
	})

	// Direct (直接攻撃) - 1枚
	attackDirect, defenseDirect := 2, 1
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-direct-assassin", prefix),
		Name:    "Shadow Assassin",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackDirect,
		Defense: &defenseDirect,
		Traits:  []entity.Trait{entity.TraitDirect},
	})

	// EffectShield (効果盾) - 1枚
	attackShield, defenseShield := 3, 4
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-shield-golem", prefix),
		Name:    "Shielded Golem",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attackShield,
		Defense: &defenseShield,
		Traits:  []entity.Trait{entity.TraitEffectShield},
	})

	// Untargetable (対象不可) - 1枚
	attackUntarget, defenseUntarget := 3, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-untarget-phantom", prefix),
		Name:    "Phantom",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackUntarget,
		Defense: &defenseUntarget,
		Traits:  []entity.Trait{entity.TraitUntargetable},
	})

	// 複数特性持ち - 1枚
	attackMulti, defenseMulti := 4, 4
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-multi-hero", prefix),
		Name:    "Heroic Defender",
		Cost:    6,
		Type:    entity.CardTypeUnit,
		Attack:  &attackMulti,
		Defense: &defenseMulti,
		Traits:  []entity.Trait{entity.TraitRush, entity.TraitGuardian},
	})

	// ========================================
	// スペルカード（16枚）CardEffectを使った複雑な効果
	// ========================================

	// 1. ダメージスペル - 敵1体に3ダメージ (2コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-fireball", prefix),
		"Fireball",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-fireball", prefix),
					Name:          "Fireball",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDealDamage,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Value:  3,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
		},
	))

	// 2. 範囲ダメージスペル - 敵全体に2ダメージ (4コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-meteor", prefix),
		"Meteor Storm",
		4,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-meteor", prefix),
					Name:          "Meteor Storm",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDealSplash,
								Target: entity.TargetSelector{Type: entity.EffectTargetEnemies},
								Value:  2,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
		},
	))

	// 3. 単体大ダメージスペル - 敵1体に5ダメージ (4コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-lightning", prefix),
		"Lightning Bolt",
		4,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-lightning", prefix),
					Name:          "Lightning Bolt",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDealDamage,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Value:  5,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
		},
	))

	// 4. 回復スペル - 自分のHPを5回復 (2コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-heal", prefix),
		Name: "Healing Light",
		Cost: 2,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-heal", prefix),
					Name:          "Healing Light",
					Description:   "自分のHPを5回復",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectRestoreHP,
								Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
								Value:  5,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "自分のHPを5回復",
		},
	})

	// 5. ドローカード - カードを2枚引く (3コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-draw", prefix),
		Name: "Arcane Wisdom",
		Cost: 3,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-draw", prefix),
					Name:          "Arcane Wisdom",
					Description:   "カードを2枚引く",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDrawCard,
								Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
								Value:  2,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "カードを2枚引く",
		},
	})

	// 6. バフスペル - 味方1体の攻撃力+3 (2コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-strengthen", prefix),
		Name: "Power Boost",
		Cost: 2,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-strengthen", prefix),
					Name:          "Power Boost",
					Description:   "味方1体の攻撃力+3",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectModifyAttack,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Value:  3,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "味方1体の攻撃力+3",
		},
	})

	// 7. 防御バフスペル - 味方1体の防御力+3 (2コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-fortify", prefix),
		Name: "Iron Skin",
		Cost: 2,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-fortify", prefix),
					Name:          "Iron Skin",
					Description:   "味方1体の防御力+3",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectModifyDefense,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Value:  3,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "味方1体の防御力+3",
		},
	})

	// 8. 全体バフスペル - 味方全体の攻撃力+2 (4コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-rally", prefix),
		Name: "Rally",
		Cost: 4,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-rally", prefix),
					Name:          "Rally",
					Description:   "味方全体の攻撃力+2",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectModifyAttack,
								Target: entity.TargetSelector{Type: entity.EffectTargetAllies},
								Value:  2,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "味方全体の攻撃力+2",
		},
	})

	// 9. 破壊スペル - 敵1体を破壊 (5コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-destroy", prefix),
		Name: "Annihilate",
		Cost: 5,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-destroy", prefix),
					Name:          "Annihilate",
					Description:   "敵1体を破壊",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDestroyUnit,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "敵1体を破壊",
		},
	})

	// 10. 手札に戻すスペル - ユニット1体を手札に戻す (3コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-bounce", prefix),
		Name: "Recall",
		Cost: 3,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-bounce", prefix),
					Name:          "Recall",
					Description:   "ユニット1体を手札に戻す",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectReturnToHand,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "ユニット1体を手札に戻す",
		},
	})

	// 11. 複合効果スペル - 3ダメージ + カード1枚引く (4コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-arcane-blast", prefix),
		"Arcane Blast",
		4,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-arcane-blast", prefix),
					Name:          "Arcane Blast",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDealDamage,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Value:  3,
								Timing: entity.EffectTimingImmediate,
							},
							Next: &entity.EffectChainNode{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectDrawCard,
										Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
										Value:  1,
										Timing: entity.EffectTimingImmediate,
									},
								},
							},
						},
					},
				},
			},
		},
	))

	// 12. 並列効果スペル - 敵1体に2ダメージ AND 味方全体に+1攻撃力 (5コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-tactical-strike", prefix),
		"Tactical Strike",
		5,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-tactical", prefix),
					Name:          "Tactical Strike",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorParallel,
						Parallel: &entity.ParallelNode{
							Children: []*entity.EffectChainNode{
								{
									Type: entity.OperatorSequential,
									Sequential: &entity.SequentialNode{
										Effect: &entity.AtomicEffect{
											Type:   entity.AtomicEffectDealDamage,
											Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
											Value:  2,
											Timing: entity.EffectTimingImmediate,
										},
									},
								},
								{
									Type: entity.OperatorSequential,
									Sequential: &entity.SequentialNode{
										Effect: &entity.AtomicEffect{
											Type:   entity.AtomicEffectModifyAttack,
											Target: entity.TargetSelector{Type: entity.EffectTargetAllies},
											Value:  1,
											Timing: entity.EffectTimingImmediate,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	))

	// 13. 特性付与スペル - 味方1体にRushを付与 (3コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-haste", prefix),
		Name: "Haste",
		Cost: 3,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-haste", prefix),
					Name:          "Haste",
					Description:   "味方1体に疾走を付与",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectGrantTrait,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Timing: entity.EffectTimingImmediate,
								Parameters: map[string]interface{}{
									"trait": entity.TraitRush,
								},
							},
						},
					},
				},
			},
			Description: "味方1体に疾走を付与",
		},
	})

	// 14. マナ回復スペル - マナを2回復 (1コスト)
	deck = append(deck, entity.Card{
		ID:   fmt.Sprintf("%s-spell-mana-potion", prefix),
		Name: "Mana Potion",
		Cost: 1,
		Type: entity.CardTypeSpell,
		CardEffect: &entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-mana-potion", prefix),
					Name:          "Mana Potion",
					Description:   "マナを2回復",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectRestoreMana,
								Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
								Value:  2,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
			Description: "マナを2回復",
		},
	})

	// 15. ForEach効果 - 味方ユニット1体につきランダムな敵に1ダメージ (5コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-chain-lightning", prefix),
		"Chain Lightning",
		5,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-chain", prefix),
					Name:          "Chain Lightning",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorForEach,
						ForEach: &entity.ForEachNode{
							Target: entity.TargetSelector{Type: entity.EffectTargetAllies},
							Effect: &entity.EffectChainNode{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectDealDamage,
										Target: entity.TargetSelector{Type: entity.EffectTargetRandomEnemy},
										Value:  1,
										Timing: entity.EffectTimingImmediate,
									},
								},
							},
						},
					},
				},
			},
		},
	))

	// 16. 条件付き効果 - HPが10以下なら敵全体に4ダメージ、そうでなければ2ダメージ (6コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-desperate-blast", prefix),
		"Desperate Blast",
		6,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-desperate", prefix),
					Name:          "Desperate Blast",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorIfElse,
						IfElse: &entity.IfElseNode{
							Condition: &entity.Condition{
								Type:     entity.ConditionPlayerHP,
								Operator: entity.OperatorLessThanOrEqual,
								Value:    10,
							},
							Then: &entity.EffectChainNode{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectDealSplash,
										Target: entity.TargetSelector{Type: entity.EffectTargetEnemies},
										Value:  4,
										Timing: entity.EffectTimingImmediate,
									},
								},
							},
							Else: &entity.EffectChainNode{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectDealSplash,
										Target: entity.TargetSelector{Type: entity.EffectTargetEnemies},
										Value:  2,
										Timing: entity.EffectTimingImmediate,
									},
								},
							},
						},
					},
				},
			},
		},
	))

	return deck
}

// JoinMatchmaking マッチメイキングに参加（サーバーストリーミング）
func (h *GameConnectHandler) JoinMatchmaking(
	ctx context.Context,
	req *connect.Request[pbv1.JoinMatchmakingRequest],
	stream *connect.ServerStream[pbv1.MatchmakingResponse],
) error {
	playerID := req.Msg.GetPlayerId()
	playerName := req.Msg.GetPlayerName()

	// マッチングキューに参加
	notifyChan := h.matchmakingService.JoinQueue(playerID, playerName)

	// 接続が切れた時にキューから退出させる
	defer h.matchmakingService.LeaveQueue(playerID)

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
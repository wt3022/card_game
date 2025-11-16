package service

import (
	"fmt"
	"sync"
	"time"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
)

// ========================================
// マッチメイキングサービス
// プレイヤーのマッチングを管理
// 設計方針:
// - キューベースのマッチング
// - 先着順でマッチング
// - リアルタイム通知
// ========================================

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

// MatchmakingService マッチメイキングサービス
type MatchmakingService struct {
	mu             sync.RWMutex
	waitingPlayers []*MatchmakingPlayer
	gameService    *GameService
	logger         port.Logger
}

// NewMatchmakingService 新しいMatchmakingServiceを作成
func NewMatchmakingService(gameService *GameService, logger port.Logger) *MatchmakingService {
	if logger == nil {
		logger = port.NewConsoleLogger()
	}
	return &MatchmakingService{
		waitingPlayers: []*MatchmakingPlayer{},
		gameService:    gameService,
		logger:         logger,
	}
}

// JoinQueue プレイヤーをマッチングキューに追加
func (s *MatchmakingService) JoinQueue(playerID, playerName string) chan *MatchResult {
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
		go s.createGame(opponent, player)
	} else {
		// 待機リストに追加
		s.waitingPlayers = append(s.waitingPlayers, player)
		s.logger.Info("⏳ 待機中のプレイヤー数: %d", len(s.waitingPlayers))
	}

	return notifyChan
}

// createGame マッチング成功後にゲームを作成
func (s *MatchmakingService) createGame(player1, player2 *MatchmakingPlayer) {
	// ゲームIDを生成
	gameID := fmt.Sprintf("game-%s-%s", player1.PlayerID, player2.PlayerID)

	// サンプルデッキを生成
	deck1 := generateSampleDeck(player1.PlayerID)
	deck2 := generateSampleDeck(player2.PlayerID)

	// ゲームを作成
	err := s.gameService.CreateGame(
		gameID,
		player1.PlayerID,
		player1.PlayerName,
		player2.PlayerID,
		player2.PlayerName,
		deck1,
		deck2,
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
func (s *MatchmakingService) LeaveQueue(playerID string) {
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
func (s *MatchmakingService) GetWaitingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.waitingPlayers)
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

func intPtr(i int) *int {
	return &i
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

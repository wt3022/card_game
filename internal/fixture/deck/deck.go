package deck

import (
	"card_game/internal/core/entity"
	"crypto/sha256"
	"encoding/hex"
)

// generateDeterministicUUID カード名から決定的なUUIDを生成
func generateDeterministicUUID(name string) string {
	// カード名のSHA256ハッシュを計算
	hash := sha256.Sum256([]byte(name))
	// 最初の16バイトを使ってUUIDを生成
	hashStr := hex.EncodeToString(hash[:16])

	// UUID v5形式に整形（8-4-4-4-12）
	return hashStr[0:8] + "-" + hashStr[8:12] + "-" + hashStr[12:16] + "-" + hashStr[16:20] + "-" + hashStr[20:32]
}

func GenerateSampleDeck() []entity.Card {
	deck := []entity.Card{}

	// ========================================
	// 通常ユニットカード（10枚）
	// コスト1-8で、マナカーブを意識したバランス配分
	// ========================================

	// 低コスト (1-2コスト): 4枚 - 序盤の展開力を重視
	// パワーカード: 1コストで2/2の高スタッツユニット
	attack1, defense1 := 2, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Elite Goblin"),
		Name:    "Elite Goblin",
		Cost:    1,
		Type:    entity.CardTypeUnit,
		Attack:  &attack1,
		Defense: &defense1,
		Traits:  []entity.Trait{},
	})

	// パワーカード: 2コストで3/3の高スタッツユニット
	attack2, defense2 := 3, 3
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("War Veteran"),
		Name:    "War Veteran",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attack2,
		Defense: &defense2,
		Traits:  []entity.Trait{},
	})

	attack3, defense3 := 3, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Archer"),
		Name:    "Archer",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attack3,
		Defense: &defense3,
		Traits:  []entity.Trait{},
	})

	attack4, defense4 := 2, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Soldier"),
		Name:    "Soldier",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attack4,
		Defense: &defense4,
		Traits:  []entity.Trait{},
	})

	// 中コスト (3-5コスト): 4枚
	attack5, defense5 := 3, 3
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Warrior"),
		Name:    "Warrior",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attack5,
		Defense: &defense5,
		Traits:  []entity.Trait{},
	})

	attack6, defense6 := 4, 3
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Knight"),
		Name:    "Knight",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attack6,
		Defense: &defense6,
		Traits:  []entity.Trait{},
	})

	attack7, defense7 := 5, 4
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Champion"),
		Name:    "Champion",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attack7,
		Defense: &defense7,
		Traits:  []entity.Trait{},
	})

	attack8, defense8 := 4, 5
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Paladin"),
		Name:    "Paladin",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attack8,
		Defense: &defense8,
		Traits:  []entity.Trait{},
	})

	// 高コスト (6-8コスト): 2枚 - フィニッシャー
	attack9, defense9 := 6, 6
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Warlord"),
		Name:    "Warlord",
		Cost:    7,
		Type:    entity.CardTypeUnit,
		Attack:  &attack9,
		Defense: &defense9,
		Traits:  []entity.Trait{},
	})

	attack10, defense10 := 8, 8
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Colossus"),
		Name:    "Colossus",
		Cost:    8,
		Type:    entity.CardTypeUnit,
		Attack:  &attack10,
		Defense: &defense10,
		Traits:  []entity.Trait{},
	})

	// ========================================
	// 特殊能力持ちユニット（14枚）
	// 多様な特性を持つユニットで戦略性を向上
	// ========================================

	// Rush (疾走) - 3枚 - 即座にアクションできる攻撃的ユニット
	// パワーカード: 2コストで3/2 Rush
	attackRush1, defenseRush1 := 3, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Alpha Wolf"),
		Name:    "Alpha Wolf",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush1,
		Defense: &defenseRush1,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	attackRush2, defenseRush2 := 3, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Rush Knight"),
		Name:    "Rush Knight",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush2,
		Defense: &defenseRush2,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	attackRush3, defenseRush3 := 4, 3
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Swift Cavalry"),
		Name:    "Swift Cavalry",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush3,
		Defense: &defenseRush3,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	// Guardian (守護) - 3枚 - 防御的戦略の要
	attackGuard1, defenseGuard1 := 1, 4
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Sentry"),
		Name:    "Sentry",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard1,
		Defense: &defenseGuard1,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	attackGuard2, defenseGuard2 := 2, 5
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Shield Guardian"),
		Name:    "Shield Guardian",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard2,
		Defense: &defenseGuard2,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	attackGuard3, defenseGuard3 := 3, 6
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Stone Wall"),
		Name:    "Stone Wall",
		Cost:    5,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard3,
		Defense: &defenseGuard3,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	// Windfury (疾風) - 2枚 - 高リスク高リターン
	attackWind1, defenseWind1 := 2, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Wind Striker"),
		Name:    "Wind Striker",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackWind1,
		Defense: &defenseWind1,
		Traits:  []entity.Trait{entity.TraitWindfury},
	})

	attackWind2, defenseWind2 := 3, 3
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Wind Master"),
		Name:    "Wind Master",
		Cost:    6,
		Type:    entity.CardTypeUnit,
		Attack:  &attackWind2,
		Defense: &defenseWind2,
		Traits:  []entity.Trait{entity.TraitWindfury},
	})

	// Pierce (貫通) - 2枚 - Guardian対策
	attackPierce1, defensePierce1 := 3, 2
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Lance Piercer"),
		Name:    "Lance Piercer",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackPierce1,
		Defense: &defensePierce1,
		Traits:  []entity.Trait{entity.TraitPierce},
	})

	attackPierce2, defensePierce2 := 5, 3
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Pierce Dragon"),
		Name:    "Pierce Dragon",
		Cost:    6,
		Type:    entity.CardTypeUnit,
		Attack:  &attackPierce2,
		Defense: &defensePierce2,
		Traits:  []entity.Trait{entity.TraitPierce},
	})

	// Charge (突進) - 1枚
	attackCharge, defenseCharge := 3, 1
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Shadow Assassin"),
		Name:    "Shadow Assassin",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attackCharge,
		Defense: &defenseCharge,
		Traits:  []entity.Trait{entity.TraitCharge},
	})

	// EffectShield (効果盾) - 1枚
	attackShield, defenseShield := 3, 4
	deck = append(deck, entity.Card{
		ID:      generateDeterministicUUID("Shielded Golem"),
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
		ID:      generateDeterministicUUID("Phantom"),
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
		ID:      generateDeterministicUUID("Heroic Defender"),
		Name:    "Heroic Defender",
		Cost:    6,
		Type:    entity.CardTypeUnit,
		Attack:  &attackMulti,
		Defense: &defenseMulti,
		Traits:  []entity.Trait{entity.TraitRush, entity.TraitGuardian},
	})

	// ========================================
	// 召喚効果持ちユニット（6枚）
	// ========================================

	// パワーカード: 2コストで2/2 + 2/2トークン召喚
	attackSummon1, defenseSummon1 := 2, 2
	recruiterEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            generateDeterministicUUID("Mass Recruit"),
				Name:          "Mass Recruit",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Elite Recruit",
								"attack":  2,
								"defense": 2,
								"cost":    2,
								"count":   1,
							},
						},
					},
				},
			},
		},
	}
	recruiterEffect.Description = recruiterEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         generateDeterministicUUID("Master Recruiter"),
		Name:       "Master Recruiter",
		Cost:       2,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon1,
		Defense:    &defenseSummon1,
		Traits:     []entity.Trait{},
		CardEffect: recruiterEffect,
		Effect:     recruiterEffect.Description,
	})

	// 召喚効果: 2/2トークンを2体召喚 (6コスト)
	attackSummon2, defenseSummon2 := 3, 3
	captainEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            generateDeterministicUUID("Rally Troops"),
				Name:          "Rally Troops",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Soldier",
								"attack":  2,
								"defense": 2,
								"cost":    2,
								"count":   2,
							},
						},
					},
				},
			},
		},
	}
	captainEffect.Description = captainEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         generateDeterministicUUID("Captain"),
		Name:       "Captain",
		Cost:       6,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon2,
		Defense:    &defenseSummon2,
		Traits:     []entity.Trait{},
		CardEffect: captainEffect,
		Effect:     captainEffect.Description,
	})

	// パワーカード: 3コストで3/3 + 3/2 Rush召喚
	attackSummon3, defenseSummon3 := 3, 3
	warcallerEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            generateDeterministicUUID("Summon Berserker"),
				Name:          "Summon Berserker",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Elite Berserker",
								"attack":  3,
								"defense": 2,
								"cost":    3,
								"count":   1,
								"traits":  []entity.Trait{entity.TraitRush},
							},
						},
					},
				},
			},
		},
	}
	warcallerEffect.Description = warcallerEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         generateDeterministicUUID("Battle Warcaller"),
		Name:       "Battle Warcaller",
		Cost:       3,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon3,
		Defense:    &defenseSummon3,
		Traits:     []entity.Trait{},
		CardEffect: warcallerEffect,
		Effect:     warcallerEffect.Description,
	})

	// パワーカード: 4コストで4/4 Rush + 3/3トークン2体召喚
	attackSummon4, defenseSummon4 := 4, 4
	commanderEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            generateDeterministicUUID("Army Deployment"),
				Name:          "Army Deployment",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Elite Soldier",
								"attack":  3,
								"defense": 3,
								"cost":    3,
								"count":   2,
							},
						},
					},
				},
			},
		},
	}
	commanderEffect.Description = commanderEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         generateDeterministicUUID("Grand Commander"),
		Name:       "Grand Commander",
		Cost:       4,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon4,
		Defense:    &defenseSummon4,
		Traits:     []entity.Trait{entity.TraitRush},
		CardEffect: commanderEffect,
		Effect:     commanderEffect.Description,
	})

	// 1/1を2体召喚 (4コスト)
	attackSummon5, defenseSummon5 := 2, 4
	nestEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            generateDeterministicUUID("Spawn Spiders"),
				Name:          "Spawn Spiders",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Spider",
								"attack":  1,
								"defense": 1,
								"cost":    1,
								"count":   2,
							},
						},
					},
				},
			},
		},
	}
	nestEffect.Description = nestEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         generateDeterministicUUID("Spider Nest"),
		Name:       "Spider Nest",
		Cost:       4,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon5,
		Defense:    &defenseSummon5,
		Traits:     []entity.Trait{},
		CardEffect: nestEffect,
		Effect:     nestEffect.Description,
	})

	// 召喚効果: Guardianを持つ1/3を2体召喚 (4コスト)
	attackSummon6, defenseSummon6 := 3, 2
	sentinelEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            generateDeterministicUUID("Summon Sentinel"),
				Name:          "Summon Sentinel",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Sentinel",
								"attack":  1,
								"defense": 3,
								"cost":    2,
								"count":   2,
								"traits":  []entity.Trait{entity.TraitGuardian},
							},
						},
					},
				},
			},
		},
	}
	sentinelEffect.Description = sentinelEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         generateDeterministicUUID("Sentinel Master"),
		Name:       "Sentinel Master",
		Cost:       4,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon6,
		Defense:    &defenseSummon6,
		Traits:     []entity.Trait{},
		CardEffect: sentinelEffect,
		Effect:     sentinelEffect.Description,
	})

	// ========================================
	// スペルカード（10枚）
	// 多様な効果で戦略性を向上させる
	// ========================================

	// パワーカード: 1コストで4ダメージ
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Blazing Fireball"),
		"Blazing Fireball",
		1,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Blazing Fireball"),
					Name:          "Blazing Fireball",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDealDamage,
								Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
								Value:  4,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
		},
	))

	// パワーカード: 3コストで3ダメージAoE
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Apocalypse"),
		"Apocalypse",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Apocalypse"),
					Name:          "Apocalypse",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectDealSplash,
								Target: entity.TargetSelector{Type: entity.EffectTargetEnemies},
								Value:  3,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
		},
	))

	// 3. 単体大ダメージスペル - 敵1体に5ダメージ (5コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Lightning Bolt"),
		"Lightning Bolt",
		5,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Lightning Bolt"),
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

	// 4. 回復スペル - 自分のHPを6回復 (3コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Healing Light"),
		"Healing Light",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Healing Light"),
					Name:          "Healing Light",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectRestoreHP,
								Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
								Value:  6,
								Timing: entity.EffectTimingImmediate,
							},
						},
					},
				},
			},
		},
	))

	// パワーカード: 2コスト Rush付与スペル
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Battle Fury"),
		"Battle Fury",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Battle Fury"),
					Name:          "Battle Fury",
					RequireTarget: true,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorParallel,
						Parallel: &entity.ParallelNode{
							Children: []*entity.EffectChainNode{
								{
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
								{
									Type: entity.OperatorSequential,
									Sequential: &entity.SequentialNode{
										Effect: &entity.AtomicEffect{
											Type:   entity.AtomicEffectGrantTrait,
											Target: entity.TargetSelector{Type: entity.EffectTargetSpecific},
											Timing: entity.EffectTimingImmediate,
											Parameters: map[string]any{
												"trait": entity.TraitRush,
											},
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

	// 6. バフスペル - 味方1体の攻撃力+3 (2コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Power Boost"),
		"Power Boost",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Power Boost"),
					Name:          "Power Boost",
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
		},
	))

	// 7. 防御バフスペル - 味方1体の防御力+3 (2コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Iron Skin"),
		"Iron Skin",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Iron Skin"),
					Name:          "Iron Skin",
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
		},
	))

	// 8. 全体バフスペル - 味方全体の攻撃力+2 (3コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Rally"),
		"Rally",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Rally"),
					Name:          "Rally",
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
		},
	))

	// パワーカード: 3コストで確定除去
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Instant Death"),
		"Instant Death",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Instant Death"),
					Name:          "Instant Death",
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
		},
	))

	// 10. 手札に戻すスペル - ユニット1体を手札に戻す (3コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Recall"),
		"Recall",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Recall"),
					Name:          "Recall",
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
		},
	))

	// 11. 召喚スペル - 2/2トークンを1体召喚 (3コスト)
	deck = append(deck, createSpellCard(
		generateDeterministicUUID("Summon Guardian"),
		"Summon Guardian",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            generateDeterministicUUID("Summon Guardian"),
					Name:          "Summon Guardian",
					RequireTarget: false,
					Root: &entity.EffectChainNode{
						Type: entity.OperatorSequential,
						Sequential: &entity.SequentialNode{
							Effect: &entity.AtomicEffect{
								Type:   entity.AtomicEffectSummonUnit,
								Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
								Timing: entity.EffectTimingImmediate,
								Parameters: map[string]any{
									"name":    "Guardian Token",
									"attack":  2,
									"defense": 2,
									"cost":    2,
									"count":   1,
									"traits":  []entity.Trait{entity.TraitGuardian},
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

func createSpellCard(id, name string, cost int, effect *entity.CardEffect) entity.Card {
	// カード効果の説明文を自動生成
	effect.Description = effect.GenerateDescription()

	return entity.Card{
		ID:         id,
		Name:       name,
		Cost:       cost,
		Type:       entity.CardTypeSpell,
		CardEffect: effect,
		Effect:     effect.Description,
	}
}

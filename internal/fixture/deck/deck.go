package deck

import (
	"fmt"

	"card_game/internal/core/entity"
)

func GenerateSampleDeck(prefix string) []entity.Card {
	deck := []entity.Card{}

	// ========================================
	// 通常ユニットカード（10枚）
	// コスト1-8で、マナカーブを意識したバランス配分
	// ========================================

	// 低コスト (1-2コスト): 4枚 - 序盤の展開力を重視
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

	attack2, defense2 := 1, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-defender", prefix),
		Name:    "Shield Bearer",
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

	attack4, defense4 := 2, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-soldier", prefix),
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
		ID:      fmt.Sprintf("%s-unit-warrior", prefix),
		Name:    "Warrior",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attack5,
		Defense: &defense5,
		Traits:  []entity.Trait{},
	})

	attack6, defense6 := 4, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-knight", prefix),
		Name:    "Knight",
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

	// 高コスト (6-8コスト): 2枚 - フィニッシャー
	attack9, defense9 := 6, 6
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-warlord", prefix),
		Name:    "Warlord",
		Cost:    7,
		Type:    entity.CardTypeUnit,
		Attack:  &attack9,
		Defense: &defense9,
		Traits:  []entity.Trait{},
	})

	attack10, defense10 := 8, 8
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-unit-colossus", prefix),
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
	attackRush1, defenseRush1 := 2, 1
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-rush-wolf", prefix),
		Name:    "Dire Wolf",
		Cost:    2,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush1,
		Defense: &defenseRush1,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	attackRush2, defenseRush2 := 3, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-rush-knight", prefix),
		Name:    "Rush Knight",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackRush2,
		Defense: &defenseRush2,
		Traits:  []entity.Trait{entity.TraitRush},
	})

	attackRush3, defenseRush3 := 4, 3
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-rush-cavalry", prefix),
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
		ID:      fmt.Sprintf("%s-guardian-sentry", prefix),
		Name:    "Sentry",
		Cost:    3,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard1,
		Defense: &defenseGuard1,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	attackGuard2, defenseGuard2 := 2, 5
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-guardian-shield", prefix),
		Name:    "Shield Guardian",
		Cost:    4,
		Type:    entity.CardTypeUnit,
		Attack:  &attackGuard2,
		Defense: &defenseGuard2,
		Traits:  []entity.Trait{entity.TraitGuardian},
	})

	attackGuard3, defenseGuard3 := 3, 6
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-guardian-wall", prefix),
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

	// Pierce (貫通) - 2枚 - Guardian対策
	attackPierce1, defensePierce1 := 3, 2
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-pierce-lancer", prefix),
		Name:    "Lance Piercer",
		Cost:    3,
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

	// Charge (突進) - 1枚
	attackCharge, defenseCharge := 3, 1
	deck = append(deck, entity.Card{
		ID:      fmt.Sprintf("%s-charge-assassin", prefix),
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
	// 召喚効果持ちユニット（6枚）
	// ========================================

	// 召喚効果: 1/1トークンを1体召喚 (3コスト)
	attackSummon1, defenseSummon1 := 2, 2
	recruiterEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            fmt.Sprintf("%s-effect-summon-recruit", prefix),
				Name:          "Recruit",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Recruit",
								"attack":  1,
								"defense": 1,
								"cost":    1,
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
		ID:         fmt.Sprintf("%s-summon-recruit", prefix),
		Name:       "Recruiter",
		Cost:       3,
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
				ID:            fmt.Sprintf("%s-effect-summon-captain", prefix),
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
		ID:         fmt.Sprintf("%s-summon-captain", prefix),
		Name:       "Captain",
		Cost:       6,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon2,
		Defense:    &defenseSummon2,
		Traits:     []entity.Trait{},
		CardEffect: captainEffect,
		Effect:     captainEffect.Description,
	})

	// 召喚効果: Rushを持つ1/1を1体召喚 (4コスト)
	attackSummon3, defenseSummon3 := 3, 3
	warcallerEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            fmt.Sprintf("%s-effect-summon-warcaller", prefix),
				Name:          "Call to Arms",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnSummon,
							Parameters: map[string]any{
								"name":    "Berserker",
								"attack":  1,
								"defense": 1,
								"cost":    1,
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
		ID:         fmt.Sprintf("%s-summon-warcaller", prefix),
		Name:       "Warcaller",
		Cost:       4,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon3,
		Defense:    &defenseSummon3,
		Traits:     []entity.Trait{},
		CardEffect: warcallerEffect,
		Effect:     warcallerEffect.Description,
	})

	// 複合効果: 自身に+2/+2バフ + 1/1トークン召喚 (5コスト)
	attackSummon4, defenseSummon4 := 3, 3
	commanderEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            fmt.Sprintf("%s-effect-summon-commander", prefix),
				Name:          "Lead by Example",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorParallel,
					Parallel: &entity.ParallelNode{
						Children: []*entity.EffectChainNode{
							{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectModifyAttack,
										Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
										Value:  2,
										Timing: entity.EffectTimingOnSummon,
									},
								},
							},
							{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectModifyDefense,
										Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
										Value:  2,
										Timing: entity.EffectTimingOnSummon,
									},
								},
							},
							{
								Type: entity.OperatorSequential,
								Sequential: &entity.SequentialNode{
									Effect: &entity.AtomicEffect{
										Type:   entity.AtomicEffectSummonUnit,
										Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
										Timing: entity.EffectTimingOnSummon,
										Parameters: map[string]any{
											"name":    "Guard",
											"attack":  1,
											"defense": 1,
											"cost":    1,
											"count":   1,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	commanderEffect.Description = commanderEffect.GenerateDescription()
	deck = append(deck, entity.Card{
		ID:         fmt.Sprintf("%s-summon-commander", prefix),
		Name:       "Battle Commander",
		Cost:       5,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon4,
		Defense:    &defenseSummon4,
		Traits:     []entity.Trait{},
		CardEffect: commanderEffect,
		Effect:     commanderEffect.Description,
	})

	// 死亡時効果: 1/1を2体召喚 (4コスト)
	attackSummon5, defenseSummon5 := 2, 4
	nestEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            fmt.Sprintf("%s-effect-summon-nest", prefix),
				Name:          "Spawn Spiders",
				RequireTarget: false,
				Root: &entity.EffectChainNode{
					Type: entity.OperatorSequential,
					Sequential: &entity.SequentialNode{
						Effect: &entity.AtomicEffect{
							Type:   entity.AtomicEffectSummonUnit,
							Target: entity.TargetSelector{Type: entity.EffectTargetSelf},
							Timing: entity.EffectTimingOnDestroy,
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
		ID:         fmt.Sprintf("%s-summon-nest", prefix),
		Name:       "Spider Nest",
		Cost:       4,
		Type:       entity.CardTypeUnit,
		Attack:     &attackSummon5,
		Defense:    &defenseSummon5,
		Traits:     []entity.Trait{},
		CardEffect: nestEffect,
		Effect:     nestEffect.Description,
	})

	// 召喚効果: Guardianを持つ1/3を1体召喚 (4コスト)
	attackSummon6, defenseSummon6 := 3, 2
	sentinelEffect := &entity.CardEffect{
		Definitions: []*entity.EffectDefinition{
			{
				ID:            fmt.Sprintf("%s-effect-summon-sentinel", prefix),
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
								"count":   1,
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
		ID:         fmt.Sprintf("%s-summon-sentinel", prefix),
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

	// 2. 範囲ダメージスペル - 敵全体に2ダメージ (5コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-meteor", prefix),
		"Meteor Storm",
		5,
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

	// 3. 単体大ダメージスペル - 敵1体に5ダメージ (5コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-lightning", prefix),
		"Lightning Bolt",
		5,
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

	// 4. 回復スペル - 自分のHPを6回復 (3コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-heal", prefix),
		"Healing Light",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-heal", prefix),
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

	// 5. ドローカード - カードを2枚引く (2コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-draw", prefix),
		"Arcane Wisdom",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-draw", prefix),
					Name:          "Arcane Wisdom",
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
		},
	))

	// 6. バフスペル - 味方1体の攻撃力+3 (2コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-strengthen", prefix),
		"Power Boost",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-strengthen", prefix),
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
		fmt.Sprintf("%s-spell-fortify", prefix),
		"Iron Skin",
		2,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-fortify", prefix),
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
		fmt.Sprintf("%s-spell-rally", prefix),
		"Rally",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-rally", prefix),
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

	// 9. 破壊スペル - 敵1体を破壊 (6コスト)
	deck = append(deck, createSpellCard(
		fmt.Sprintf("%s-spell-destroy", prefix),
		"Annihilate",
		6,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-destroy", prefix),
					Name:          "Annihilate",
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
		fmt.Sprintf("%s-spell-bounce", prefix),
		"Recall",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-bounce", prefix),
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
		fmt.Sprintf("%s-spell-summon-token", prefix),
		"Summon Guardian",
		3,
		&entity.CardEffect{
			Definitions: []*entity.EffectDefinition{
				{
					ID:            fmt.Sprintf("%s-effect-summon-token", prefix),
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

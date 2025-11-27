package converter

import (
	cardgamev1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/internal/core/entity"
	"card_game/internal/core/usecase/game"
	"card_game/internal/infrastructure/event"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ========================================
// Entity → Proto 変換
// ドメインエンティティをProtocol Buffersメッセージに変換
// 設計方針:
// - 情報隠蔽: 相手の手札内容は送信しない
// - Nil安全: nilチェックを徹底
// - 完全なマッピング: すべてのフィールドを変換
// ========================================

// GameStateToProto ゲーム状態をProtoに変換
func GameStateToProto(state *game.State, viewerPlayerID string) *cardgamev1.GameState {
	if state == nil {
		return nil
	}

	player1 := state.GetPlayerByID(state.Player1.ID)
	player2 := state.GetPlayerByID(state.Player2.ID)

	return &cardgamev1.GameState{
		GameId:              state.GameID,
		Player1:             PlayerToProto(player1, viewerPlayerID),
		Player2:             PlayerToProto(player2, viewerPlayerID),
		CurrentPlayerId:     state.CurrentPlayerID,
		CurrentTurn:         int32(state.CurrentTurn),
		CurrentPhase:        GamePhaseToProto(state.CurrentPhase),
		IsGameOver:          state.IsOver(),
		WinnerId:            stringPtrToOptional(state.WinnerID),
		IsDraw:              state.WinnerID != nil && *state.WinnerID == "",
		Player1MulliganDone: state.Player1MulliganDone,
		Player2MulliganDone: state.Player2MulliganDone,
		CoinTossDone:        state.CoinTossDone,
		CoinTossWinnerId:    stringPtrToOptional(state.CoinTossWinnerID),
		TurnOrderDecided:    state.TurnOrderDecided,
	}
}

// PlayerToProto プレイヤー情報をProtoに変換
func PlayerToProto(player *entity.Player, viewerPlayerID string) *cardgamev1.Player {
	if player == nil {
		return nil
	}

	// 情報隠蔽: 自分のプレイヤーの場合のみ手札の詳細を含める
	var hand []*cardgamev1.Card
	if player.ID == viewerPlayerID {
		hand = CardsToProto(player.Hand)
	}

	return &cardgamev1.Player{
		Id:                   player.ID,
		Name:                 player.Name,
		Hp:                   int32(player.HP),
		MaxHp:                int32(player.MaxHP),
		CurrentTurnMana:      int32(player.CurrentTurnMana),
		CurrentRecoveryMana:  int32(player.CurrentRecoveryMana),
		HandCount:            int32(len(player.Hand)),
		DeckCount:            int32(len(player.Deck)),
		GraveyardCount:       int32(len(player.Graveyard)),
		Field:                UnitsToProto(player.Field),
		Hand:                 hand,
		TimeRemainingSeconds: 0, // ターン時間制限機能は将来実装予定
		IsConnected:          player.IsConnected,
		LastActivityAt:       timestamppb.New(player.LastActivityAt),
	}
}

// UnitToProto ユニットをProtoに変換
func UnitToProto(unit *entity.Unit) *cardgamev1.Unit {
	if unit == nil {
		return nil
	}

	return &cardgamev1.Unit{
		CardId:           unit.CardID,
		InstanceId:       unit.InstanceID,
		Name:             unit.Name,
		Cost:             int32(unit.Cost),
		Attack:           int32(unit.Attack),
		Defense:          int32(unit.Defense),
		CurrentDefense:   int32(unit.CurrentDefense),
		Traits:           TraitsToProto(unit.Traits),
		Effect:           unit.Effect,
		AttacksRemaining: int32(unit.AttacksRemaining),
		SummonedThisTurn: unit.SummonedThisTurn,
		OwnerId:          unit.OwnerID,
	}
}

// UnitsToProto ユニットリストをProtoに変換
func UnitsToProto(units []entity.Unit) []*cardgamev1.Unit {
	result := make([]*cardgamev1.Unit, len(units))
	for i, unit := range units {
		unitCopy := unit // コピーを作成
		result[i] = UnitToProto(&unitCopy)
	}
	return result
}

// CardToProto カードをProtoに変換
func CardToProto(card *entity.Card) *cardgamev1.Card {
	if card == nil {
		return nil
	}

	// 表示用テキスト: Card.Effect が空なら CardEffect.Description を優先して使用
	effectText := card.Effect
	if effectText == "" && card.CardEffect != nil {
		effectText = card.CardEffect.Description
	}

	// CardEffectの変換
	// Note: entity.CardEffectとproto CardEffectは設計が異なる
	// entity側: ゲームエンジン用の実行可能な構造体
	// proto側: 管理UI用のJSON化可能な構造
	// CardEffectはhandler層で直接repository経由で取得して設定される
	var cardEffect *cardgamev1.CardEffect
	// ここではnilのまま。handler層でGetCardEffectAsProtoを使って設定される

	return &cardgamev1.Card{
		Id:         card.ID,
		Name:       card.Name,
		Type:       CardTypeToProto(card.Type),
		Cost:       int32(card.Cost),
		Attack:     intPtrToOptionalInt32(card.Attack),
		Defense:    intPtrToOptionalInt32(card.Defense),
		Effect:     effectText,
		Traits:     TraitsToProto(card.Traits),
		CardEffect: cardEffect, // handler層で設定される
	}
}

// CardsToProto カードリストをProtoに変換
func CardsToProto(cards []entity.Card) []*cardgamev1.Card {
	result := make([]*cardgamev1.Card, len(cards))
	for i, card := range cards {
		cardCopy := card
		result[i] = CardToProto(&cardCopy)
	}
	return result
}

// GameEventToProto ゲームイベントをProtoに変換
func GameEventToProto(event *event.GameEvent) *cardgamev1.GameEvent {
	if event == nil {
		return nil
	}

	return &cardgamev1.GameEvent{
		GameId:    event.GameID,
		Turn:      int32(event.State.CurrentTurn),
		Phase:     string(event.State.CurrentPhase),
		EventType: event.EventType,
		PlayerId:  event.PlayerID,
		Details:   event.Message,
		Timestamp: timestamppb.New(event.Timestamp),
	}
}

// CombatResultToProto 戦闘結果をProtoに変換
func CombatResultToProto(result *entity.CombatResult) *cardgamev1.AttackResult {
	if result == nil {
		return &cardgamev1.AttackResult{}
	}

	return &cardgamev1.AttackResult{
		AttackerDestroyed: result.AttackerDestroyed,
		DefenderDestroyed: result.DefenderDestroyed,
		Damage:            int32(result.Damage),
		DirectDamage:      int32(result.DirectDamage),
	}
}

// ========================================
// Enum変換
// ========================================

// CardTypeToProto CardType変換
func CardTypeToProto(cardType entity.CardType) cardgamev1.CardType {
	switch cardType {
	case entity.CardTypeUnit:
		return cardgamev1.CardType_CARD_TYPE_UNIT
	case entity.CardTypeSpell:
		return cardgamev1.CardType_CARD_TYPE_SPELL
	case entity.CardTypeLeader:
		return cardgamev1.CardType_CARD_TYPE_LEADER
	default:
		return cardgamev1.CardType_CARD_TYPE_UNSPECIFIED
	}
}

// TraitToProto Trait変換
func TraitToProto(trait entity.Trait) cardgamev1.Trait {
	switch trait {
	case entity.TraitRush:
		return cardgamev1.Trait_TRAIT_RUSH
	case entity.TraitCharge:
		return cardgamev1.Trait_TRAIT_CHARGE
	case entity.TraitWindfury:
		return cardgamev1.Trait_TRAIT_WINDFURY
	case entity.TraitPierce:
		return cardgamev1.Trait_TRAIT_PIERCE
	case entity.TraitGuardian:
		return cardgamev1.Trait_TRAIT_GUARDIAN
	case entity.TraitEffectShield:
		return cardgamev1.Trait_TRAIT_EFFECT_SHIELD
	case entity.TraitUntargetable:
		return cardgamev1.Trait_TRAIT_UNTARGETABLE
	default:
		return cardgamev1.Trait_TRAIT_UNSPECIFIED
	}
}

// TraitsToProto Traitリスト変換
func TraitsToProto(traits []entity.Trait) []cardgamev1.Trait {
	result := make([]cardgamev1.Trait, len(traits))
	for i, trait := range traits {
		result[i] = TraitToProto(trait)
	}
	return result
}

// GamePhaseToProto GamePhase変換
func GamePhaseToProto(phase entity.GamePhase) cardgamev1.GamePhase {
	switch phase {
	case entity.GamePhaseTurnStart:
		return cardgamev1.GamePhase_GAME_PHASE_TURN_START
	case entity.GamePhaseDraw:
		return cardgamev1.GamePhase_GAME_PHASE_DRAW
	case entity.GamePhaseResourceGain:
		return cardgamev1.GamePhase_GAME_PHASE_RESOURCE_GAIN
	case entity.GamePhaseMain:
		return cardgamev1.GamePhase_GAME_PHASE_MAIN
	case entity.GamePhaseTurnEnd:
		return cardgamev1.GamePhase_GAME_PHASE_TURN_END
	default:
		return cardgamev1.GamePhase_GAME_PHASE_UNSPECIFIED
	}
}

// ========================================
// ヘルパー関数
// ========================================

// intPtrToOptionalInt32 *int を optional int32 に変換
func intPtrToOptionalInt32(v *int) *int32 {
	if v == nil {
		return nil
	}
	val := int32(*v)
	return &val
}

// stringPtrToOptional *string を optional string に変換
func stringPtrToOptional(v *string) *string {
	if v == nil {
		return nil
	}
	return v
}

// DeckToProto デッキをProtoに変換
func DeckToProto(deck *entity.Deck) *cardgamev1.Deck {
	if deck == nil {
		return nil
	}

	return &cardgamev1.Deck{
		Id:          deck.ID,
		Name:        deck.Name,
		Description: deck.Description,
		CardIds:     deck.CardIDs,
		UserId:      deck.UserID,
		CreatedAt:   timestamppb.New(deck.CreatedAt),
		UpdatedAt:   timestamppb.New(deck.UpdatedAt),
	}
}

// ========================================
// CardEffect変換
// ========================================

// CardEffectToProto CardEffectをProtoに変換
func CardEffectToProto(cardEffect *entity.CardEffect) *cardgamev1.CardEffect {
	if cardEffect == nil {
		return nil
	}

	definitions := make([]*cardgamev1.EffectDefinition, len(cardEffect.Definitions))
	for i, def := range cardEffect.Definitions {
		definitions[i] = effectDefinitionToProto(def)
	}

	return &cardgamev1.CardEffect{
		CardId:      "", // CardIDはhandlerで設定
		Definitions: definitions,
	}
}

func effectDefinitionToProto(def *entity.EffectDefinition) *cardgamev1.EffectDefinition {
	if def == nil {
		return nil
	}

	return &cardgamev1.EffectDefinition{
		Id:            uint32(0), // IDはDBで管理
		RequireTarget: def.RequireTarget,
		Root:          effectChainNodeToProto(def.Root),
	}
}

func effectChainNodeToProto(node *entity.EffectChainNode) *cardgamev1.EffectChainNode {
	if node == nil {
		return nil
	}

	pbNode := &cardgamev1.EffectChainNode{
		Id:   uint32(0), // IDはDBで管理
		Type: effectChainNodeTypeToProto(node.Type),
	}

	switch node.Type {
	case entity.OperatorSequential:
		if node.Sequential != nil {
			pbNode.AtomicEffect = atomicEffectToProto(node.Sequential.Effect)
			pbNode.Next = effectChainNodeToProto(node.Sequential.Next)
		}
	case entity.OperatorParallel:
		if node.Parallel != nil {
			children := make([]*cardgamev1.EffectChainNode, len(node.Parallel.Children))
			for i, child := range node.Parallel.Children {
				children[i] = effectChainNodeToProto(child)
			}
			pbNode.Children = children
		}
	case entity.OperatorIfElse:
		if node.IfElse != nil {
			pbNode.Condition = conditionToProto(node.IfElse.Condition)
			pbNode.ThenNode = effectChainNodeToProto(node.IfElse.Then)
			pbNode.ElseNode = effectChainNodeToProto(node.IfElse.Else)
		}
	case entity.OperatorRepeat:
		if node.Repeat != nil {
			count := int32(node.Repeat.Count)
			pbNode.RepeatCount = &count
			pbNode.RepeatEffect = effectChainNodeToProto(node.Repeat.Effect)
		}
	case entity.OperatorForEach:
		if node.ForEach != nil {
			pbNode.ForeachTarget = targetSelectorToProto(&node.ForEach.Target)
			pbNode.ForeachEffect = effectChainNodeToProto(node.ForEach.Effect)
		}
	}

	return pbNode
}

func effectChainNodeTypeToProto(t entity.EffectOperator) cardgamev1.EffectChainNodeType {
	switch t {
	case entity.OperatorSequential:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_THEN
	case entity.OperatorParallel:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_AND
	case entity.OperatorIfElse:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_IF_ELSE
	case entity.OperatorRepeat:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_REPEAT
	case entity.OperatorForEach:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_FOREACH
	default:
		return cardgamev1.EffectChainNodeType_EFFECT_CHAIN_NODE_TYPE_UNSPECIFIED
	}
}

func atomicEffectToProto(effect *entity.AtomicEffect) *cardgamev1.AtomicEffect {
	if effect == nil {
		return nil
	}

	pbEffect := &cardgamev1.AtomicEffect{
		Id:     uint32(0), // IDはDBで管理
		Type:   atomicEffectTypeToProto(effect.Type),
		Target: targetSelectorToProto(&effect.Target),
		Timing: effectTimingToProto(effect.Timing),
	}

	if effect.Value != 0 {
		value := int32(effect.Value)
		pbEffect.Value = &value
	}

	if cardID, ok := effect.Parameters["card_id"].(string); ok {
		pbEffect.CardId = &cardID
	}

	if trait, ok := effect.Parameters["trait"].(entity.Trait); ok {
		pbTrait := TraitToProto(trait)
		pbEffect.Trait = &pbTrait
	}

	// 全てのparametersをStructとして設定
	if len(effect.Parameters) > 0 {
		pbEffect.Parameters = convertMapToStruct(effect.Parameters)
	}

	return pbEffect
}

func atomicEffectTypeToProto(t entity.AtomicEffectType) cardgamev1.AtomicEffectType {
	switch t {
	case entity.AtomicEffectDealDamage:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_DAMAGE
	case entity.AtomicEffectDealSplash:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DEAL_SPLASH
	case entity.AtomicEffectRestoreHP:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_HP
	case entity.AtomicEffectRestoreMana:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RESTORE_MANA
	case entity.AtomicEffectFullRestore:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_FULL_RESTORE
	case entity.AtomicEffectDrawCard:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DRAW_CARD
	case entity.AtomicEffectDiscardCard:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISCARD_CARD
	case entity.AtomicEffectSearchCard:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SEARCH_CARD
	case entity.AtomicEffectShuffleDeck:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SHUFFLE_DECK
	case entity.AtomicEffectModifyAttack:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_ATTACK
	case entity.AtomicEffectModifyDefense:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_DEFENSE
	case entity.AtomicEffectModifyCost:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_COST
	case entity.AtomicEffectModifyMaxHP:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_MODIFY_MAX_HP
	case entity.AtomicEffectSummonUnit:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_SUMMON_UNIT
	case entity.AtomicEffectDestroyUnit:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DESTROY_UNIT
	case entity.AtomicEffectReturnToHand:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_HAND
	case entity.AtomicEffectReturnToDeck:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_RETURN_TO_DECK
	case entity.AtomicEffectSilenceUnit:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_DISABLE_UNIT
	case entity.AtomicEffectGrantTrait:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GRANT_TRAIT
	case entity.AtomicEffectRemoveTrait:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REMOVE_TRAIT
	case entity.AtomicEffectGainMana:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_GAIN_MANA
	case entity.AtomicEffectReduceCost:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_REDUCE_COST
	default:
		return cardgamev1.AtomicEffectType_ATOMIC_EFFECT_TYPE_UNSPECIFIED
	}
}

func effectTimingToProto(timing entity.EffectTiming) cardgamev1.EffectTiming {
	switch timing {
	case entity.EffectTimingImmediate:
		return cardgamev1.EffectTiming_EFFECT_TIMING_IMMEDIATE
	case entity.EffectTimingOnSummon:
		return cardgamev1.EffectTiming_EFFECT_TIMING_ON_SUMMON
	case entity.EffectTimingOnDestroy:
		return cardgamev1.EffectTiming_EFFECT_TIMING_ON_DESTROY
	case entity.EffectTimingOnAttack:
		return cardgamev1.EffectTiming_EFFECT_TIMING_ON_ATTACK
	case entity.EffectTimingOnDamaged:
		return cardgamev1.EffectTiming_EFFECT_TIMING_ON_DAMAGED
	case entity.EffectTimingTurnStart:
		return cardgamev1.EffectTiming_EFFECT_TIMING_TURN_START
	case entity.EffectTimingTurnEnd:
		return cardgamev1.EffectTiming_EFFECT_TIMING_TURN_END
	default:
		return cardgamev1.EffectTiming_EFFECT_TIMING_UNSPECIFIED
	}
}

func targetSelectorToProto(selector *entity.TargetSelector) *cardgamev1.TargetSelector {
	if selector == nil {
		return nil
	}

	return &cardgamev1.TargetSelector{
		Id:     uint32(0), // IDはDBで管理
		Type:   targetTypeToProto(selector.Type),
		Filter: targetFilterToProto(selector.Filter),
	}
}

func targetTypeToProto(t entity.EffectTarget) cardgamev1.TargetType {
	switch t {
	case entity.EffectTargetSelf:
		return cardgamev1.TargetType_TARGET_TYPE_SELF
	case entity.EffectTargetOpponent:
		return cardgamev1.TargetType_TARGET_TYPE_ENEMY_LEADER
	case entity.EffectTargetAllies:
		return cardgamev1.TargetType_TARGET_TYPE_ALL_ALLY_UNITS
	case entity.EffectTargetEnemies:
		return cardgamev1.TargetType_TARGET_TYPE_ALL_ENEMY_UNITS
	case entity.EffectTargetAllUnits:
		return cardgamev1.TargetType_TARGET_TYPE_ALL_UNITS
	case entity.EffectTargetRandomAlly:
		return cardgamev1.TargetType_TARGET_TYPE_RANDOM_ALLY_UNIT
	case entity.EffectTargetRandomEnemy:
		return cardgamev1.TargetType_TARGET_TYPE_RANDOM_ENEMY_UNIT
	case entity.EffectTargetSpecific:
		return cardgamev1.TargetType_TARGET_TYPE_UNSPECIFIED
	default:
		return cardgamev1.TargetType_TARGET_TYPE_UNSPECIFIED
	}
}

func conditionToProto(cond *entity.Condition) *cardgamev1.ConditionFilter {
	if cond == nil {
		return nil
	}

	// TODO: Conditionの適切な変換ロジックを実装
	return &cardgamev1.ConditionFilter{
		Id:            uint32(0),
		ConditionType: string(cond.Type),
		Parameters:    []string{},
	}
}

func targetFilterToProto(filter *entity.TargetFilter) *cardgamev1.ConditionFilter {
	if filter == nil {
		return nil
	}

	// TODO: TargetFilterの適切な変換ロジックを実装
	return &cardgamev1.ConditionFilter{
		Id:            uint32(0),
		ConditionType: "",
		Parameters:    []string{},
	}
}

// convertMapToStruct map[string]any を google.protobuf.Struct に変換
func convertMapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		return nil
	}

	// map[string]any を structpb に変換
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}

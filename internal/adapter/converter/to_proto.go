package converter

import (
	"encoding/json"

	cardgamev1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/internal/core/entity"
	"card_game/internal/core/usecase/game"
	"card_game/internal/infrastructure/event"

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
		TimeRemainingSeconds: 0, // TODO: 将来実装
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

	var cardEffectJSON string
	if card.CardEffect != nil {
		if data, err := json.Marshal(card.CardEffect); err == nil {
			cardEffectJSON = string(data)
		}
	}

	return &cardgamev1.Card{
		Id:             card.ID,
		Name:           card.Name,
		Type:           CardTypeToProto(card.Type),
		Cost:           int32(card.Cost),
		Attack:         intPtrToOptionalInt32(card.Attack),
		Defense:        intPtrToOptionalInt32(card.Defense),
		Effect:         effectText,
		Traits:         TraitsToProto(card.Traits),
		CardEffectJson: cardEffectJSON,
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

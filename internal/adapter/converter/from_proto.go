package converter

import (
	cardgamev1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/internal/core/entity"
)

// ========================================
// Proto → Entity 変換
// Protocol Buffersメッセージをドメインエンティティに変換
// 設計方針:
// - バリデーション: 不正な値は適切にエラー処理
// - Nil安全: nilチェックを徹底
// - デフォルト値: UNSPECIFIED値を適切に処理
// ========================================

// CardFromProto ProtoからCardに変換
func CardFromProto(pb *cardgamev1.Card) *entity.Card {
	if pb == nil {
		return nil
	}

	return &entity.Card{
		ID:      pb.Id,
		Name:    pb.Name,
		Type:    CardTypeFromProto(pb.Type),
		Cost:    int(pb.Cost),
		Attack:  optionalInt32ToPtrInt(pb.Attack),
		Defense: optionalInt32ToPtrInt(pb.Defense),
		Effect:  pb.Effect,
		Traits:  TraitsFromProto(pb.Traits),
		// CardEffectは設定しない（サーバー側でのみ管理）
	}
}

// UnitFromProto ProtoからUnitに変換
func UnitFromProto(pb *cardgamev1.Unit) *entity.Unit {
	if pb == nil {
		return nil
	}

	return &entity.Unit{
		CardID:           pb.CardId,
		InstanceID:       pb.InstanceId,
		Name:             pb.Name,
		Cost:             int(pb.Cost),
		Attack:           int(pb.Attack),
		Defense:          int(pb.Defense),
		CurrentDefense:   int(pb.CurrentDefense),
		Traits:           TraitsFromProto(pb.Traits),
		Effect:           pb.Effect,
		AttacksRemaining: int(pb.AttacksRemaining),
		SummonedThisTurn: pb.SummonedThisTurn,
		OwnerID:          pb.OwnerId,
	}
}

// ========================================
// Enum変換
// ========================================

// CardTypeFromProto CardType変換
func CardTypeFromProto(pbType cardgamev1.CardType) entity.CardType {
	switch pbType {
	case cardgamev1.CardType_CARD_TYPE_UNIT:
		return entity.CardTypeUnit
	case cardgamev1.CardType_CARD_TYPE_SPELL:
		return entity.CardTypeSpell
	case cardgamev1.CardType_CARD_TYPE_LEADER:
		return entity.CardTypeLeader
	default:
		return entity.CardTypeUnit // デフォルト
	}
}

// TraitFromProto Trait変換
func TraitFromProto(pbTrait cardgamev1.Trait) entity.Trait {
	switch pbTrait {
	case cardgamev1.Trait_TRAIT_RUSH:
		return entity.TraitRush
	case cardgamev1.Trait_TRAIT_CHARGE:
		return entity.TraitCharge
	case cardgamev1.Trait_TRAIT_WINDFURY:
		return entity.TraitWindfury
	case cardgamev1.Trait_TRAIT_PIERCE:
		return entity.TraitPierce
	case cardgamev1.Trait_TRAIT_GUARDIAN:
		return entity.TraitGuardian
	case cardgamev1.Trait_TRAIT_EFFECT_SHIELD:
		return entity.TraitEffectShield
	case cardgamev1.Trait_TRAIT_UNTARGETABLE:
		return entity.TraitUntargetable
	default:
		return "" // 不明な特性は空文字列
	}
}

// TraitsFromProto Traitリスト変換
func TraitsFromProto(pbTraits []cardgamev1.Trait) []entity.Trait {
	result := make([]entity.Trait, 0, len(pbTraits))
	for _, pbTrait := range pbTraits {
		trait := TraitFromProto(pbTrait)
		if trait != "" {
			result = append(result, trait)
		}
	}
	return result
}

// GamePhaseFromProto GamePhase変換
func GamePhaseFromProto(pbPhase cardgamev1.GamePhase) entity.GamePhase {
	switch pbPhase {
	case cardgamev1.GamePhase_GAME_PHASE_TURN_START:
		return entity.GamePhaseTurnStart
	case cardgamev1.GamePhase_GAME_PHASE_DRAW:
		return entity.GamePhaseDraw
	case cardgamev1.GamePhase_GAME_PHASE_RESOURCE_GAIN:
		return entity.GamePhaseResourceGain
	case cardgamev1.GamePhase_GAME_PHASE_MAIN:
		return entity.GamePhaseMain
	case cardgamev1.GamePhase_GAME_PHASE_TURN_END:
		return entity.GamePhaseTurnEnd
	default:
		return entity.GamePhaseMain // デフォルト
	}
}

// ========================================
// ヘルパー関数
// ========================================

// optionalInt32ToPtrInt optional int32 を *int に変換
func optionalInt32ToPtrInt(v *int32) *int {
	if v == nil {
		return nil
	}
	val := int(*v)
	return &val
}

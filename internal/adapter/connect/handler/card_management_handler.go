package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/converter"
	"card_game/internal/application/service"
	"card_game/internal/core/entity"
	"card_game/internal/infrastructure/persistence/model"
)

// ========================================
// カード管理ハンドラー
// ========================================

// CardManagementConnectHandler Connect-Go用のカード管理サービスハンドラー
type CardManagementConnectHandler struct {
	cardService *service.CardService
}

// NewCardManagementConnectHandler 新しいCardManagementConnectHandlerを作成
func NewCardManagementConnectHandler(cardService *service.CardService) *CardManagementConnectHandler {
	return &CardManagementConnectHandler{
		cardService: cardService,
	}
}

// インターフェースの実装確認
var _ cardgamev1connect.CardManagementServiceHandler = (*CardManagementConnectHandler)(nil)

// CreateCard カードを作成
func (h *CardManagementConnectHandler) CreateCard(
	ctx context.Context,
	req *connect.Request[pbv1.CreateCardRequest],
) (*connect.Response[pbv1.CreateCardResponse], error) {
	// protoからエンティティに変換
	card, err := h.protoToCard(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// カードを作成
	if err := h.cardService.CreateCard(card); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// CardEffectをmodel構造で保存
	if req.Msg.GetCardEffectJson() != "" {
		cardEffectModel := &model.CardEffectModel{}
		if err := json.Unmarshal([]byte(req.Msg.GetCardEffectJson()), cardEffectModel); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to unmarshal card effect: %w", err))
		}
		if err := h.cardService.SaveCardEffect(card.ID, cardEffectModel); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// エンティティからprotoに変換
	protoCard := converter.CardToProto(card)

	resp := &pbv1.CreateCardResponse{
		Card: protoCard,
	}

	return connect.NewResponse(resp), nil
}

// GetCard カードを取得
func (h *CardManagementConnectHandler) GetCard(
	ctx context.Context,
	req *connect.Request[pbv1.GetCardRequest],
) (*connect.Response[pbv1.GetCardResponse], error) {
	cardID := req.Msg.GetId()

	// カードを取得
	card, err := h.cardService.GetCard(cardID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// エンティティからprotoに変換
	protoCard := converter.CardToProto(card)

	resp := &pbv1.GetCardResponse{
		Card: protoCard,
	}

	return connect.NewResponse(resp), nil
}

// ListCards カード一覧を取得
func (h *CardManagementConnectHandler) ListCards(
	ctx context.Context,
	req *connect.Request[pbv1.ListCardsRequest],
) (*connect.Response[pbv1.ListCardsResponse], error) {
	var cards []*entity.Card
	var err error

	// タイプが指定されている場合はフィルタリング
	if req.Msg.Type != nil && *req.Msg.Type != pbv1.CardType_CARD_TYPE_UNSPECIFIED {
		cardType := converter.CardTypeFromProto(*req.Msg.Type)
		cards, err = h.cardService.ListCardsByType(cardType)
	} else {
		cards, err = h.cardService.ListCards()
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// エンティティからprotoに変換
	protoCards := make([]*pbv1.Card, len(cards))
	for i, card := range cards {
		protoCards[i] = converter.CardToProto(card)
	}

	resp := &pbv1.ListCardsResponse{
		Cards: protoCards,
	}

	return connect.NewResponse(resp), nil
}

// UpdateCard カードを更新
func (h *CardManagementConnectHandler) UpdateCard(
	ctx context.Context,
	req *connect.Request[pbv1.UpdateCardRequest],
) (*connect.Response[pbv1.UpdateCardResponse], error) {
	// protoからエンティティに変換
	card, err := h.protoToCardForUpdate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// カードを更新
	if err := h.cardService.UpdateCard(card); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// CardEffectをmodel構造で保存
	if req.Msg.GetCardEffectJson() != "" {
		cardEffectModel := &model.CardEffectModel{}
		if err := json.Unmarshal([]byte(req.Msg.GetCardEffectJson()), cardEffectModel); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to unmarshal card effect: %w", err))
		}
		if err := h.cardService.SaveCardEffect(card.ID, cardEffectModel); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// エンティティからprotoに変換
	protoCard := converter.CardToProto(card)

	resp := &pbv1.UpdateCardResponse{
		Card: protoCard,
	}

	return connect.NewResponse(resp), nil
}

// DeleteCard カードを削除
func (h *CardManagementConnectHandler) DeleteCard(
	ctx context.Context,
	req *connect.Request[pbv1.DeleteCardRequest],
) (*connect.Response[pbv1.DeleteCardResponse], error) {
	cardID := req.Msg.GetId()

	// カードを削除
	if err := h.cardService.DeleteCard(cardID); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	resp := &pbv1.DeleteCardResponse{
		Success: true,
		Message: "Card deleted successfully",
	}

	return connect.NewResponse(resp), nil
}

// protoToCard protoからCardエンティティに変換（作成用）
func (h *CardManagementConnectHandler) protoToCard(msg *pbv1.CreateCardRequest) (*entity.Card, error) {
	cardType := converter.CardTypeFromProto(msg.GetType())
	traits := converter.TraitsFromProto(msg.GetTraits())

	// スペルまたはリーダーの場合は特性をクリア
	if cardType == entity.CardTypeSpell || cardType == entity.CardTypeLeader {
		traits = []entity.Trait{}
	}

	card := &entity.Card{
		ID:     msg.GetId(),
		Name:   msg.GetName(),
		Type:   cardType,
		Cost:   int(msg.GetCost()),
		Effect: "", // 効果テキストは自動生成されるため空文字列を保存
		Traits: traits,
	}

	if msg.Attack != nil {
		attack := int(*msg.Attack)
		card.Attack = &attack
	}

	if msg.Defense != nil {
		defense := int(*msg.Defense)
		card.Defense = &defense
	}

	// CardEffectをJSONから復元（model構造に直接Unmarshal）
	if msg.GetCardEffectJson() != "" {
		cardEffectModel := &model.CardEffectModel{}
		if err := json.Unmarshal([]byte(msg.GetCardEffectJson()), cardEffectModel); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card effect: %w", err)
		}
		// modelからentityに変換（将来実装）
		// 現在はmodel構造を直接使用
		card.CardEffect = nil // TODO: modelからentityに変換
	}

	return card, nil
}

// protoToCardForUpdate protoからCardエンティティに変換（更新用）
func (h *CardManagementConnectHandler) protoToCardForUpdate(msg *pbv1.UpdateCardRequest) (*entity.Card, error) {
	cardType := converter.CardTypeFromProto(msg.GetType())
	traits := converter.TraitsFromProto(msg.GetTraits())

	// スペルまたはリーダーの場合は特性をクリア
	if cardType == entity.CardTypeSpell || cardType == entity.CardTypeLeader {
		traits = []entity.Trait{}
	}

	card := &entity.Card{
		ID:     msg.GetId(),
		Name:   msg.GetName(),
		Type:   cardType,
		Cost:   int(msg.GetCost()),
		Effect: "", // 効果テキストは自動生成されるため空文字列を保存
		Traits: traits,
	}

	if msg.Attack != nil {
		attack := int(*msg.Attack)
		card.Attack = &attack
	}

	if msg.Defense != nil {
		defense := int(*msg.Defense)
		card.Defense = &defense
	}

	// CardEffectをJSONから復元（model構造に直接Unmarshal）
	if msg.GetCardEffectJson() != "" {
		cardEffectModel := &model.CardEffectModel{}
		if err := json.Unmarshal([]byte(msg.GetCardEffectJson()), cardEffectModel); err != nil {
			return nil, fmt.Errorf("failed to unmarshal card effect: %w", err)
		}
		// modelからentityに変換（将来実装）
		// 現在はmodel構造を直接使用
		card.CardEffect = nil // TODO: modelからentityに変換
	}

	return card, nil
}

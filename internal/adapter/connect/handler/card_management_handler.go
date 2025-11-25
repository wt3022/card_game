package handler

import (
	"context"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/converter"
	"card_game/internal/application/service"
	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/repository"
)

// ========================================
// カード管理ハンドラー
// ========================================

// CardManagementConnectHandler Connect-Go用のカード管理サービスハンドラー
type CardManagementConnectHandler struct {
	cardService *service.CardService
	deckService *service.DeckService
	cardRepo    port.CardRepository
}

// NewCardManagementConnectHandler 新しいCardManagementConnectHandlerを作成
func NewCardManagementConnectHandler(cardService *service.CardService, deckService *service.DeckService, cardRepo port.CardRepository) *CardManagementConnectHandler {
	return &CardManagementConnectHandler{
		cardService: cardService,
		deckService: deckService,
		cardRepo:    cardRepo,
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

	// CardEffectを保存(protoから変換)
	if req.Msg.CardEffect != nil {
		cardEffectModel, err := repository.CardEffectFromProtoToModel(req.Msg.CardEffect)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if err := h.cardRepo.SaveCardEffect(card.ID, cardEffectModel); err != nil {
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

	// CardEffectをProto形式でロード
	cardEffectProto, err := h.loadAndConvertCardEffect(cardID)
	if err == nil && cardEffectProto != nil {
		protoCard.CardEffect = cardEffectProto
	}

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
		// CardEffectをProto形式でロード
		cardEffectProto, err := h.loadAndConvertCardEffect(card.ID)
		if err == nil && cardEffectProto != nil {
			protoCards[i].CardEffect = cardEffectProto
		}
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

	// CardEffectを保存(protoから変換)
	if req.Msg.CardEffect != nil {
		cardEffectModel, err := repository.CardEffectFromProtoToModel(req.Msg.CardEffect)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if err := h.cardRepo.SaveCardEffect(card.ID, cardEffectModel); err != nil {
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

	// CardEffectはハンドラーで直接Modelに変換して保存するため、ここでは設定しない

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

	// CardEffectはハンドラーで直接Modelに変換して保存するため、ここでは設定しない

	return card, nil
}

// loadAndConvertCardEffect CardEffectModelをロードしてProtoに変換
func (h *CardManagementConnectHandler) loadAndConvertCardEffect(cardID string) (*pbv1.CardEffect, error) {
	cardEffectProtoInterface, err := h.cardRepo.GetCardEffectAsProto(cardID)
	if err != nil {
		return nil, err
	}
	if cardEffectProtoInterface == nil {
		return nil, nil
	}

	// 型アサーション
	cardEffectProto, ok := cardEffectProtoInterface.(*pbv1.CardEffect)
	if !ok {
		return nil, nil
	}

	return cardEffectProto, nil
}

// ========================================
// デッキ管理エンドポイント
// ========================================

// CreateDeck デッキを作成
func (h *CardManagementConnectHandler) CreateDeck(
	ctx context.Context,
	req *connect.Request[pbv1.CreateDeckRequest],
) (*connect.Response[pbv1.CreateDeckResponse], error) {
	// 認証情報からユーザーIDを取得
	userID := "admin" // TODO: Interceptorから認証情報を取得する実装に変更

	deck, err := h.deckService.CreateDeck(ctx, req.Msg.Name, req.Msg.Description, userID, req.Msg.CardIds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	deckProto := converter.DeckToProto(deck)
	return connect.NewResponse(&pbv1.CreateDeckResponse{
		Deck: deckProto,
	}), nil
}

// GetDeck デッキを取得
func (h *CardManagementConnectHandler) GetDeck(
	ctx context.Context,
	req *connect.Request[pbv1.GetDeckRequest],
) (*connect.Response[pbv1.GetDeckResponse], error) {
	deck, err := h.deckService.GetDeck(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	deckProto := converter.DeckToProto(deck)
	return connect.NewResponse(&pbv1.GetDeckResponse{
		Deck: deckProto,
	}), nil
}

// ListDecks デッキ一覧を取得
func (h *CardManagementConnectHandler) ListDecks(
	ctx context.Context,
	req *connect.Request[pbv1.ListDecksRequest],
) (*connect.Response[pbv1.ListDecksResponse], error) {
	var decks []*entity.Deck
	var err error

	if req.Msg.UserId != nil && *req.Msg.UserId != "" {
		decks, err = h.deckService.ListDecksByUser(ctx, *req.Msg.UserId)
	} else {
		decks, err = h.deckService.ListAllDecks(ctx)
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	deckProtos := make([]*pbv1.Deck, len(decks))
	for i, deck := range decks {
		deckProtos[i] = converter.DeckToProto(deck)
	}

	return connect.NewResponse(&pbv1.ListDecksResponse{
		Decks: deckProtos,
	}), nil
}

// UpdateDeck デッキを更新
func (h *CardManagementConnectHandler) UpdateDeck(
	ctx context.Context,
	req *connect.Request[pbv1.UpdateDeckRequest],
) (*connect.Response[pbv1.UpdateDeckResponse], error) {
	deck, err := h.deckService.UpdateDeck(ctx, req.Msg.Id, req.Msg.Name, req.Msg.Description, req.Msg.CardIds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	deckProto := converter.DeckToProto(deck)
	return connect.NewResponse(&pbv1.UpdateDeckResponse{
		Deck: deckProto,
	}), nil
}

// DeleteDeck デッキを削除
func (h *CardManagementConnectHandler) DeleteDeck(
	ctx context.Context,
	req *connect.Request[pbv1.DeleteDeckRequest],
) (*connect.Response[pbv1.DeleteDeckResponse], error) {
	if err := h.deckService.DeleteDeck(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pbv1.DeleteDeckResponse{
		Success: true,
		Message: "Deck deleted successfully",
	}), nil
}

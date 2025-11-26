package handler

import (
	"context"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/converter"
	"card_game/internal/application/service"
	"card_game/internal/core/entity"
)

// ========================================
// カード管理ハンドラー
// ========================================

// getUserIDFromContext コンテキストからユーザーIDを取得
// Note: 現状はInterceptorで"user_id"キーにユーザーIDが設定されることを前提とする
// 開発環境ではDEV_MODE環境変数が"true"の場合のみ"admin"をデフォルトとして返す
func getUserIDFromContext(ctx context.Context) string {
	// コンテキストから取得を試みる
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		return userID
	}
	// 開発環境でのみデフォルト値を使用（本番環境では空文字列を返す）
	// TODO: 本番環境ではエラーを返すべき
	// if os.Getenv("DEV_MODE") == "true" {
	// 	return "admin"
	// }
	// return ""

	// 現状は互換性のため常に"admin"を返す（将来的に上記のロジックに置き換える）
	return "admin"
}

// CardManagementConnectHandler Connect-Go用のカード管理サービスハンドラー
type CardManagementConnectHandler struct {
	cardService *service.CardService
	deckService *service.DeckService
}

// NewCardManagementConnectHandler 新しいCardManagementConnectHandlerを作成
func NewCardManagementConnectHandler(cardService *service.CardService, deckService *service.DeckService) *CardManagementConnectHandler {
	return &CardManagementConnectHandler{
		cardService: cardService,
		deckService: deckService,
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

	// CardEffectを保存
	if req.Msg.CardEffect != nil {
		// ProtoからEntityに変換
		cardEffect := converter.CardEffectFromProto(req.Msg.CardEffect)
		if err := h.cardService.SaveCardEffect(card.ID, cardEffect); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}

		// 効果テキストを自動生成してカードに設定
		effectDescription, err := h.cardService.GenerateEffectDescription(card.ID)
		if err == nil && effectDescription != "" {
			card.Effect = effectDescription
			// カードを再度更新して効果テキストを保存
			if err := h.cardService.UpdateCard(card); err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
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

	// CardEffectを取得してProtoに変換
	cardEffect, err := h.cardService.GetCardEffect(cardID)
	if err == nil && cardEffect != nil {
		protoCard.CardEffect = converter.CardEffectToProto(cardEffect)
		if protoCard.CardEffect != nil {
			protoCard.CardEffect.CardId = cardID
		}
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
		// CardEffectを取得してProtoに変換
		cardEffect, err := h.cardService.GetCardEffect(card.ID)
		if err == nil && cardEffect != nil {
			protoCards[i].CardEffect = converter.CardEffectToProto(cardEffect)
			if protoCards[i].CardEffect != nil {
				protoCards[i].CardEffect.CardId = card.ID
			}
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

	// CardEffectを先に保存して効果テキストを生成
	if req.Msg.CardEffect != nil {
		// ProtoからEntityに変換
		cardEffect := converter.CardEffectFromProto(req.Msg.CardEffect)
		if err := h.cardService.SaveCardEffect(card.ID, cardEffect); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}

		// CardEffectから効果テキストを自動生成してカードに設定（上書き）
		effectDescription, err := h.cardService.GenerateEffectDescription(card.ID)
		if err == nil && effectDescription != "" {
			card.Effect = effectDescription
		}
	}

	// カードを更新（効果テキストを含む）
	if err := h.cardService.UpdateCard(card); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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

// ========================================
// デッキ管理エンドポイント
// ========================================

// CreateDeck デッキを作成
func (h *CardManagementConnectHandler) CreateDeck(
	ctx context.Context,
	req *connect.Request[pbv1.CreateDeckRequest],
) (*connect.Response[pbv1.CreateDeckResponse], error) {
	// 認証情報からユーザーIDを取得
	userID := getUserIDFromContext(ctx)

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

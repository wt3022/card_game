package handler

import (
	"context"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/application/service"
)

// ========================================
// 認証ハンドラー
// ========================================

// AuthConnectHandler Connect-Go用の認証サービスハンドラー
type AuthConnectHandler struct {
	authService *service.AuthService
}

// NewAuthConnectHandler 新しいAuthConnectHandlerを作成
func NewAuthConnectHandler(authService *service.AuthService) *AuthConnectHandler {
	return &AuthConnectHandler{
		authService: authService,
	}
}

// インターフェースの実装確認
var _ cardgamev1connect.AuthServiceHandler = (*AuthConnectHandler)(nil)

// Login ログイン処理
func (h *AuthConnectHandler) Login(
	ctx context.Context,
	req *connect.Request[pbv1.LoginRequest],
) (*connect.Response[pbv1.LoginResponse], error) {
	username := req.Msg.GetUsername()
	password := req.Msg.GetPassword()

	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			connect.NewError(connect.CodeInvalidArgument, nil))
	}

	// 認証サービスでログイン
	loginResp, err := h.authService.Login(username, password)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// レスポンスを作成
	resp := &pbv1.LoginResponse{
		AccessToken: loginResp.AccessToken,
		UserId:      loginResp.UserID,
		Username:    loginResp.Username,
		Role:        loginResp.Role,
	}

	return connect.NewResponse(resp), nil
}

// RefreshToken トークンリフレッシュ（簡易実装、将来拡張）
func (h *AuthConnectHandler) RefreshToken(
	ctx context.Context,
	req *connect.Request[pbv1.RefreshTokenRequest],
) (*connect.Response[pbv1.RefreshTokenResponse], error) {
	// 現在は未実装（将来拡張用）
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

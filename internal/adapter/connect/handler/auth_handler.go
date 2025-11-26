package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	pbv1 "card_game/api/gen/proto/cardgame/v1"
	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/application/service"
	"card_game/internal/core/port"
)

// ========================================
// 認証ハンドラー
// ========================================

// AuthConnectHandler Connect-Go用の認証サービスハンドラー
type AuthConnectHandler struct {
	authService *service.AuthService
	logger      port.Logger
}

// NewAuthConnectHandler 新しいAuthConnectHandlerを作成
func NewAuthConnectHandler(authService *service.AuthService, logger port.Logger) *AuthConnectHandler {
	return &AuthConnectHandler{
		authService: authService,
		logger:      logger,
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

	h.logger.Info("Login attempt for user: %s", username)

	if username == "" || password == "" {
		h.logger.Error("Login failed: empty username or password")
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("ユーザー名とパスワードは必須です"))
	}

	// 認証サービスでログイン
	loginResp, err := h.authService.Login(username, password)
	if err != nil {
		h.logger.Error("Login failed for user '%s': %v", username, err)
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	h.logger.Info("Login successful for user: %s", username)

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

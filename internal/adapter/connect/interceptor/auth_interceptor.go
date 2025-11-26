package interceptor

import (
	"context"
	"strings"

	"connectrpc.com/connect"

	"card_game/internal/core/port"
)

// ========================================
// 認証インターセプター
// ========================================

// AuthInterceptor 認証を検証するインターセプター
type AuthInterceptor struct {
	tokenProvider port.TokenProvider
	logger        port.Logger
}

// NewAuthInterceptor 新しい認証インターセプターを作成
func NewAuthInterceptor(tokenProvider port.TokenProvider, logger port.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		tokenProvider: tokenProvider,
		logger:        logger,
	}
}

// WrapUnary ユニタリーRPC用の認証インターセプター
func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		i.logger.Info("Auth interceptor: checking request to %s", req.Spec().Procedure)

		// Authorizationヘッダーからトークンを取得
		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			i.logger.Error("Auth interceptor: no Authorization header")
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		// "Bearer "プレフィックスを除去
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// Bearerプレフィックスがない場合
			i.logger.Error("Auth interceptor: no Bearer prefix")
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		i.logger.Info("Auth interceptor: validating token (length: %d)", len(tokenString))

		// トークンを検証
		claims, err := i.tokenProvider.ValidateToken(tokenString)
		if err != nil {
			i.logger.Error("Auth interceptor: token validation failed: %v", err)
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		i.logger.Info("Auth interceptor: token valid for user %s", claims.Username)

		// コンテキストにクレームを追加
		ctx = context.WithValue(ctx, "jwt_claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)

		// 次のハンドラーを呼び出し
		return next(ctx, req)
	}
}

// WrapStreamingClient ストリーミングクライアント用（未実装）
func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler ストリーミングハンドラー用
func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// Authorizationヘッダーからトークンを取得
		authHeader := conn.RequestHeader().Get("Authorization")
		if authHeader == "" {
			return connect.NewError(connect.CodeUnauthenticated, nil)
		}

		// "Bearer "プレフィックスを除去
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return connect.NewError(connect.CodeUnauthenticated, nil)
		}

		// トークンを検証
		claims, err := i.tokenProvider.ValidateToken(tokenString)
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated, err)
		}

		// コンテキストにクレームを追加
		ctx = context.WithValue(ctx, "jwt_claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)

		// 次のハンドラーを呼び出し
		return next(ctx, conn)
	}
}

// NewAuthInterceptorFunc Connect-RPCのUnaryInterceptorFuncとして使用する関数
func NewAuthInterceptorFunc(tokenProvider port.TokenProvider, logger port.Logger) connect.UnaryInterceptorFunc {
	interceptor := NewAuthInterceptor(tokenProvider, logger)
	return interceptor.WrapUnary
}

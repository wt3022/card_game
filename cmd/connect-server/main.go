package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/connect/handler"
	"card_game/internal/adapter/connect/interceptor"
	"card_game/internal/application/service"
	"card_game/internal/infrastructure/auth"
	"card_game/internal/infrastructure/logger"
	"card_game/internal/infrastructure/middleware"
	"card_game/internal/infrastructure/persistence"
	"card_game/internal/infrastructure/repository"
	"card_game/internal/infrastructure/security"

	"connectrpc.com/connect"
	"github.com/joho/godotenv"
)

// ========================================
// Connect-Goサーバー
// protoベースのHTTP/JSON + gRPC通信
// 設計方針:
// - プロキシ不要（直接ブラウザと通信可能）
// - HTTP/1.1とHTTP/2の両方をサポート
// - CORS対応
// - JWT認証対応
// ========================================

func main() {
	// .envファイルを読み込む（存在しない場合はスキップ）
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .envファイルが見つかりません。環境変数またはデフォルト値を使用します")
	}

	// ロガーを初期化
	logger := logger.NewConsoleLogger()

	// データベース接続
	dbConfig := persistence.NewDBConfig()

	// SQL接続（マイグレーション用）
	sqlDB, err := persistence.OpenDB(dbConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer sqlDB.Close()
	log.Println("✅ データベース接続成功")

	// GORM接続（アプリケーション用）
	gormDB, err := persistence.OpenGormDB(dbConfig)
	if err != nil {
		log.Fatalf("GORM データベース接続に失敗しました: %v", err)
	}
	log.Println("✅ GORM データベース接続成功")

	// GORMマイグレーションを実行
	if err := persistence.RunGormMigrations(gormDB, logger); err != nil {
		log.Fatalf("GORM マイグレーション実行に失敗しました: %v", err)
	}
	log.Println("✅ GORM マイグレーション完了")

	// JWTプロバイダーを初期化
	tokenProvider, err := auth.NewJWTProvider()
	if err != nil {
		log.Fatalf("JWTプロバイダーの初期化に失敗しました: %v", err)
	}

	// パスワードハッシャーを初期化
	passwordHasher := auth.NewPasswordHasher()

	// リポジトリを初期化(GORMを使用)
	userRepo := repository.NewUserRepository(gormDB)
	cardRepo := repository.NewCardRepository(gormDB)
	deckRepo := repository.NewDeckRepository(gormDB)

	// 初期管理者ユーザーを作成(存在しない場合)
	if err := service.InitializeDefaultAdmin(userRepo, passwordHasher, logger); err != nil {
		log.Printf("⚠️  デフォルト管理者ユーザーの初期化に失敗しました: %v", err)
	}

	// セキュリティコンポーネントを初期化
	rateLimiter := middleware.NewRateLimiter()
	rateLimiter.StartCleanupRoutine()
	stateValidator := security.NewStateValidator()
	cheatDetector := security.NewCheatDetector()
	log.Println("✅ セキュリティコンポーネント初期化完了")

	// サービスを初期化
	authService := service.NewAuthService(userRepo, tokenProvider, passwordHasher, logger)
	cardService := service.NewCardService(cardRepo, logger)
	deckService := service.NewDeckService(deckRepo, cardRepo, logger)
	gameService := service.NewGameServiceWithSecurity(deckService, logger, rateLimiter, stateValidator, cheatDetector)

	// 認証インターセプターを作成
	authInterceptor := interceptor.NewAuthInterceptorFunc(tokenProvider, logger)

	// Connect-Goハンドラーを初期化
	gameHandler := handler.NewGameConnectHandler(gameService)
	authHandler := handler.NewAuthConnectHandler(authService, logger)
	cardManagementHandler := handler.NewCardManagementConnectHandler(cardService, deckService)

	// マルチプレクサを作成
	mux := http.NewServeMux()

	// 認証不要のエンドポイント（AuthService）
	authPath, authHandlerFunc := cardgamev1connect.NewAuthServiceHandler(
		authHandler,
	)
	mux.Handle(authPath, authHandlerFunc)

	// 認証不要のエンドポイント（GameService）
	gamePath, gameHandlerFunc := cardgamev1connect.NewGameServiceHandler(
		gameHandler,
	)
	mux.Handle(gamePath, gameHandlerFunc)

	// 認証必須のエンドポイント（CardManagementService）
	cardMgmtPath, cardMgmtHandlerFunc := cardgamev1connect.NewCardManagementServiceHandler(
		cardManagementHandler,
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(cardMgmtPath, cardMgmtHandlerFunc)

	// ヘルスチェックエンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// CORS対応ミドルウェア（ハードコードで許可）
	httpHandler := corsMiddleware(mux)

	// HTTP/2をサポート（h2c = HTTP/2 Cleartext）
	h2cHandler := h2c.NewHandler(httpHandler, &http2.Server{})

	// サーバー起動
	port := getEnv("GAME_SERVER_PORT", "8080")
	addr := fmt.Sprintf(":%s", port)

	log.Printf("🎮 Connect-Go サーバー起動: http://localhost%s", addr)
	log.Printf("📡 gRPC-Web & HTTP/JSON エンドポイント:")
	log.Printf("   %s (認証 - 認証不要)", authPath)
	log.Printf("   %s (ゲーム - 認証不要)", gamePath)
	log.Printf("   %s (カード管理 - 認証必須)", cardMgmtPath)
	log.Printf("💡 特徴:")
	log.Printf("   - プロキシ不要（直接ブラウザと通信）")
	log.Printf("   - HTTP/1.1 + HTTP/2対応")
	log.Printf("   - JSON & Protocol Buffers対応")
	log.Printf("   - CORS対応")
	log.Printf("   - JWT認証対応")

	if err := http.ListenAndServe(addr, h2cHandler); err != nil {
		log.Fatalf("サーバー起動に失敗しました: %v", err)
	}
}

// corsMiddleware CORS対応ミドルウェア
func corsMiddleware(next http.Handler) http.Handler {
	// 許可するオリジンのホワイトリスト
	allowedOrigins := map[string]bool{
		"https://www.release-notifier.net": true,
		"https://api.release-notifier.net": true,
		"http://localhost:3000":            true, // 開発用
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// オリジンがホワイトリストに含まれているかチェック
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// 許可されていないオリジンの場合は明示的に拒否
			w.Header().Set("Access-Control-Allow-Origin", "null")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version, Connect-Timeout-Ms")

		// Private Network Access (PNA) を許可するヘッダ
		// Chrome/Edge等で https → localhost への通信に必要
		w.Header().Set("Access-Control-Allow-Private-Network", "true")

		// セキュリティヘッダー
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Strict-Transport-Security は HTTPS 環境でのみ有効
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// プリフライトリクエストの処理
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getEnv 環境変数を取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

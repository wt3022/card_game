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
	"card_game/internal/adapter/jwt"
	"card_game/internal/application/service"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence"
	"card_game/internal/infrastructure/repository"

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
		log.Println("⚠️  .env file not found, using environment variables or defaults")
	}

	// ロガーを初期化
	logger := port.NewConsoleLogger()

	// データベース接続
	dbConfig := persistence.NewDBConfig()

	// SQL接続（マイグレーション用）
	sqlDB, err := persistence.OpenDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlDB.Close()
	log.Println("✅ Database connected")

	// GORM接続（アプリケーション用）
	gormDB, err := persistence.OpenGormDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database with GORM: %v", err)
	}
	log.Println("✅ GORM Database connected")

	// GORMマイグレーションを実行
	if err := persistence.RunGormMigrations(gormDB, logger); err != nil {
		log.Fatalf("Failed to run GORM migrations: %v", err)
	}
	log.Println("✅ GORM Migrations completed")

	// JWTプロバイダーを初期化
	tokenProvider, err := jwt.NewJWTProvider()
	if err != nil {
		log.Fatalf("Failed to initialize JWT provider: %v", err)
	}

	// パスワードハッシャーを初期化
	passwordHasher := jwt.NewPasswordHasher()

	// リポジトリを初期化（GORMを使用）
	userRepo := repository.NewUserRepository(gormDB)
	cardRepo := repository.NewCardRepository(gormDB)

	// 初期管理者ユーザーを作成（存在しない場合）
	if err := service.InitializeDefaultAdmin(userRepo, passwordHasher, logger); err != nil {
		log.Printf("⚠️  Failed to initialize default admin user: %v", err)
		// エラーでも続行（既に存在する場合など）
	}

	// サービスを初期化
	gameService := service.NewGameService(logger)
	authService := service.NewAuthService(userRepo, tokenProvider, passwordHasher, logger)
	cardService := service.NewCardService(cardRepo, logger)

	// 認証インターセプターを作成
	authInterceptor := interceptor.NewAuthInterceptorFunc(tokenProvider)

	// Connect-Goハンドラーを初期化
	gameHandler := handler.NewGameConnectHandler(gameService)
	authHandler := handler.NewAuthConnectHandler(authService)
	cardManagementHandler := handler.NewCardManagementConnectHandler(cardService)

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

	log.Printf("🎮 Connect-Go Server starting on http://localhost%s", addr)
	log.Printf("📡 gRPC-Web & HTTP/JSON endpoints:")
	log.Printf("   %s (Auth - No auth required)", authPath)
	log.Printf("   %s (Game - No auth required)", gamePath)
	log.Printf("   %s (Card Management - Auth required)", cardMgmtPath)
	log.Printf("💡 特徴:")
	log.Printf("   - プロキシ不要（直接ブラウザと通信）")
	log.Printf("   - HTTP/1.1 + HTTP/2対応")
	log.Printf("   - JSON & Protocol Buffers対応")
	log.Printf("   - CORS対応")
	log.Printf("   - JWT認証対応")

	if err := http.ListenAndServe(addr, h2cHandler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware CORS対応ミドルウェア
func corsMiddleware(next http.Handler) http.Handler {
	// 許可するオリジンのホワイトリスト
	allowedOrigins := map[string]bool{
		"https://www.release-notifier.net": true,
		"https://api.release-notifier.net": true,
		"http://localhost:3000":            true, // 開発用
		"http://localhost:5173":            true, // Vite開発サーバー用
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

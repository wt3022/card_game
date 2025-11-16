package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"card_game/api/gen/proto/cardgame/v1/cardgamev1connect"
	"card_game/internal/adapter/connect/handler"
	"card_game/internal/application/service"
	"card_game/internal/core/port"
)

// ========================================
// Connect-Goサーバー
// protoベースのHTTP/JSON + gRPC通信
// 設計方針:
// - プロキシ不要（直接ブラウザと通信可能）
// - HTTP/1.1とHTTP/2の両方をサポート
// - CORS対応
// ========================================

func main() {
	// ロガーを初期化
	logger := port.NewConsoleLogger()

	// GameServiceを初期化
	gameService := service.NewGameService(logger)

	// MatchmakingServiceを初期化
	matchmakingService := service.NewMatchmakingService(gameService, logger)

	// Connect-Goハンドラーを初期化
	connectHandler := handler.NewGameConnectHandler(gameService, matchmakingService)

	// マルチプレクサを作成
	mux := http.NewServeMux()

	// Connect-Goのパスを登録
	path, handlerFunc := cardgamev1connect.NewGameServiceHandler(connectHandler)
	mux.Handle(path, handlerFunc)

	// ヘルスチェックエンドポイント
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// CORS対応ミドルウェア（ハードコードで許可）
	handler := corsMiddleware(mux)

	// HTTP/2をサポート（h2c = HTTP/2 Cleartext）
	h2cHandler := h2c.NewHandler(handler, &http2.Server{})

	// サーバー起動
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	log.Printf("🎮 Connect-Go Server starting on http://localhost%s", addr)
	log.Printf("📡 gRPC-Web & HTTP/JSON endpoints:")
	log.Printf("   %s", path)
	log.Printf("💡 特徴:")
	log.Printf("   - プロキシ不要（直接ブラウザと通信）")
	log.Printf("   - HTTP/1.1 + HTTP/2対応")
	log.Printf("   - JSON & Protocol Buffers対応")
	log.Printf("   - CORS対応")
	log.Printf("")
	log.Printf("🧪 テスト方法:")
	log.Printf("   curl -X POST http://localhost%s/cardgame.v1.GameService/CreateGame \\", addr)
	log.Printf("     -H 'Content-Type: application/json' \\")
	log.Printf("     -d '{\"player1_id\":\"p1\",\"player1_name\":\"Alice\",\"player2_id\":\"p2\",\"player2_name\":\"Bob\"}'")

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
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
.PHONY: help proto proto-go proto-connect install-tools build-server build-connect clean test

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ========================================
# プロトコルバッファ関連
# ========================================

install-tools: ## 必要なツールをインストール
	@echo "📦 Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@echo "✅ Tools installed successfully"

proto: proto-connect ## すべてのprotoファイルからコードを生成

proto-connect: ## Connect-Go用のコードを生成
	@echo "🔨 Generating Connect-Go code from proto files..."
	protoc -I=api/proto \
		--go_out=api/gen/proto/cardgame/v1 --go_opt=paths=source_relative \
		--connect-go_out=api/gen/proto/cardgame/v1 --connect-go_opt=paths=source_relative \
		api/proto/*.proto
	@echo "✅ Connect-Go code generated"

# ========================================
# ビルド関連
# ========================================

build-connect: ## Connect-Goサーバーをビルド
	@echo "🔨 Building Connect-Go server..."
	go build -o bin/connect-server cmd/connect-server/main.go
	@echo "✅ Connect-Go server built: bin/connect-server"


# ========================================
# 実行関連
# ========================================

run-connect: ## Connect-Goサーバーを起動
	@echo "🚀 Starting Connect-Go server..."
	go run cmd/connect-server/main.go

# ========================================
# その他
# ========================================

clean: ## 生成ファイルとビルド成果物を削除
	@echo "🧹 Cleaning generated files..."
	rm -rf api/gen/
	rm -rf bin/
	@echo "✅ Clean complete"

test: ## テストを実行
	@echo "🧪 Running tests..."
	go test -v ./...

deps: ## 依存関係を更新
	@echo "📦 Updating dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies updated"

# ========================================
# 開発用
# ========================================

dev: proto build-connect run-connect ## 開発用: proto生成→ビルド→実行


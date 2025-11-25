.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Proto生成
.PHONY: proto
proto: ## Protoファイルからコードを生成
	@echo "🔨 Generating proto files..."
	protoc --proto_path=api/proto \
		--go_out=api/gen/proto/cardgame/v1 --go_opt=paths=source_relative \
		--go-grpc_out=api/gen/proto/cardgame/v1 --go-grpc_opt=paths=source_relative \
		--connect-go_out=api/gen/proto/cardgame/v1 --connect-go_opt=paths=source_relative \
		api/proto/auth.proto \
		api/proto/common.proto \
		api/proto/card_management.proto \
		api/proto/game.proto


# 依存関係
.PHONY: deps
deps: ## Go依存関係をインストール
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# ツールのインストール
.PHONY: install-tools
install-tools: ## Protoツールをインストール
	@echo "🔧 Installing tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

# ビルド
.PHONY: build-connect
build-connect: ## Connect-Goサーバーをビルド
	@echo "🔨 Building Connect-Go server..."
	go build -o bin/connect-server cmd/connect-server/main.go

.PHONY: build
build: build-connect ## すべてのサーバーをビルド

# 実行
.PHONY: run-connect
run-connect: ## Connect-Goサーバーを起動
	@echo "🚀 Running Connect-Go server..."
	go run cmd/connect-server/main.go


# テスト
.PHONY: test
test: ## テストを実行
	@echo "🧪 Running tests..."
	go test -v ./...

# クリーンアップ
.PHONY: clean
clean: ## ビルド成果物を削除
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -rf api/gen/

# デフォルトターゲット
.DEFAULT_GOAL := help

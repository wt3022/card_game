.PHONY: help
help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Proto生成
.PHONY: proto
proto: ## Protoファイルからコードを生成
	@echo "🔨 Generating proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--connect-go_out=. --connect-go_opt=paths=source_relative \
		api/proto/*.proto

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

.PHONY: build-admin
build-admin: ## Admin serverをビルド
	@echo "🔨 Building Admin server..."
	go build -o bin/admin-server cmd/admin-server/main.go

.PHONY: build-migrate
build-migrate: ## Migration toolをビルド
	@echo "🔨 Building Migration tool..."
	go build -o bin/migrate cmd/migrate/main.go

.PHONY: build
build: build-connect build-admin build-migrate ## すべてのサーバーをビルド

# 実行
.PHONY: run-connect
run-connect: ## Connect-Goサーバーを起動
	@echo "🚀 Running Connect-Go server..."
	go run cmd/connect-server/main.go

.PHONY: run-admin
run-admin: ## Admin serverを起動
	@echo "🚀 Running Admin server..."
	go run cmd/admin-server/main.go

# データベース
.PHONY: migrate
migrate: ## DBマイグレーションを実行
	@echo "🗄️  Running database migration..."
	go run cmd/migrate/main.go

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

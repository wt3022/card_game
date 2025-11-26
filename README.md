# Card Game

デジタルカードバトルゲーム - Connect-RPC + React によるリアルタイムカードゲーム

## 概要

Go + React で実装されたターン制のデジタルカードゲームです。Connect-RPC を使用してバックエンドとフロントエンドが通信し、Protocol Buffers によって型安全な API を実現しています。クリーンアーキテクチャを採用し、保守性と拡張性を重視した設計となっています。

## 技術スタック

### バックエンド
- **言語**: Go 1.25.4
- **RPC**: Connect-RPC (connectrpc.com/connect)
- **認証**: JWT (golang-jwt/jwt)
- **データベース**: MySQL 8.4
- **ORM**: GORM
- **通信**: Protocol Buffers + HTTP/1.1 + HTTP/2

### フロントエンド
- **フレームワーク**: React 18 + TypeScript
- **ビルドツール**: Vite 5
- **ルーティング**: React Router v7
- **RPC クライアント**: @connectrpc/connect-web
- **コード品質**: Biome (Linter + Formatter)
- **型定義**: Protocol Buffers (自動生成)

### アーキテクチャ
- **設計パターン**: クリーンアーキテクチャ (Clean Architecture)
- **レイヤー構成**: Core / Application / Adapter / Infrastructure
- **API 定義**: Protocol Buffers v3
- **通信方式**: unary RPC + server streaming

## セットアップ

### 必要なツール

- **Go**: 1.25 以上
- **Node.js**: 18 以上
- **pnpm**: 最新版
- **Docker**: Docker Compose 対応版（MySQL 用）
- **protoc**: Protocol Buffers コンパイラ
- **buf**: Protocol Buffers ツール（フロントエンド生成用）

### 環境変数の設定

```bash
# example.env をコピーして .env を作成
cp example.env .env

# 必要に応じて .env を編集
# - MYSQL_ROOT_PASSWORD
# - MYSQL_DATABASE
# - MYSQL_PORT
# - JWT_SECRET
```

### インストール

```bash
# 1. バックエンドの依存関係をインストール
make deps

# 2. Protocol Buffers ツールをインストール
make install-tools

# 3. Protocol Buffers からコードを生成
make proto

# 4. フロントエンドの依存関係をインストール
cd frontend
pnpm install

# 5. フロントエンドの Protocol Buffers コードを生成
pnpm proto:generate
cd ..

# 6. MySQL を起動
docker-compose up -d
```

## 実行方法

### 開発環境

```bash
# 1. MySQL が起動していることを確認
docker-compose ps

# 2. バックエンドサーバーを起動
make run-connect
# => http://localhost:8080 で起動

# 3. フロントエンド開発サーバーを起動（別ターミナル）
cd frontend
pnpm dev
# => http://localhost:5173 で起動
```

- **バックエンド**: http://localhost:8080
- **フロントエンド**: http://localhost:5173

### ビルド

```bash
# バックエンドのビルド
make build-connect
# => bin/connect-server が生成される

# フロントエンドのビルド
cd frontend
pnpm build
# => dist/ ディレクトリに出力される
```

### プロダクション実行

```bash
# バックエンド
./bin/connect-server

# フロントエンド（静的ファイルサーバーが必要）
cd frontend/dist
# nginx や serve などで配信
```

## ゲームの基本ルール

### 基本仕様
- **初期 HP**: 20
- **初期マナ**: 1
- **初期手札**: 4枚（マリガン可能）
- **最大マナ**: 10
- **デッキ**: 30枚

### ゲームフロー
1. **マリガン**: 初期手札から任意のカードを引き直し
2. **ドローフェーズ**: カードを1枚ドロー（初手プレイヤーは1ターン目スキップ）
3. **メインフェーズ**: マナを使ってカードをプレイ
   - **ユニットカード**: フィールドに召喚（最大7体）
   - **スペルカード**: 即座に効果を発動
4. **バトルフェーズ**: ユニットで攻撃
   - 相手プレイヤーまたは相手ユニットを攻撃
   - **Taunt（挑発）**: Taunt を持つユニットがいる場合、そのユニット以外を攻撃不可
   - **Rush（速攻）**: 召喚ターンから攻撃可能（通常は召喚酔い）
5. **エンドフェーズ**: ターン終了
   - マナが全回復し、最大マナが +1（上限10）

### 勝利条件
- 相手プレイヤーの HP を 0 にする
- 相手がデッキ切れ（ドローできない）

## プロジェクト構成

```
card_game/
├── api/                           # API 定義
│   ├── proto/                     # Protocol Buffers 定義ファイル
│   │   ├── auth.proto            # 認証 API
│   │   ├── card_management.proto # カード管理 API
│   │   ├── game.proto            # ゲーム API
│   │   └── common.proto          # 共通型定義
│   └── gen/                       # 自動生成コード（Go）
├── cmd/
│   └── connect-server/            # Connect-RPC サーバーのエントリポイント
│       └── main.go
├── internal/                      # 内部パッケージ（クリーンアーキテクチャ）
│   ├── core/                      # コア層（ドメイン層）
│   │   ├── entity/               # エンティティ（カード、プレイヤー、ユニット等）
│   │   ├── usecase/              # ユースケース（ゲームロジック）
│   │   └── port/                 # インターフェース定義
│   ├── application/               # アプリケーション層
│   │   └── service/              # アプリケーションサービス
│   │       ├── auth_service.go   # 認証サービス
│   │       ├── card_service.go   # カード管理サービス
│   │       ├── deck_service.go   # デッキ管理サービス
│   │       └── game_service.go   # ゲームサービス
│   ├── adapter/                   # アダプター層
│   │   ├── connect/              # Connect-RPC ハンドラー
│   │   │   ├── handler/          # API ハンドラー
│   │   │   └── interceptor/      # 認証インターセプター
│   │   ├── converter/            # DTO 変換（proto ⇔ entity）
│   │   └── jwt/                  # JWT トークン処理
│   └── infrastructure/            # インフラ層
│       ├── auth/                 # 認証基盤
│       ├── event/                # イベント管理
│       ├── logger/               # ロガー
│       ├── persistence/          # DB 接続
│       └── repository/           # データアクセス
├── frontend/                      # フロントエンド
│   ├── src/
│   │   ├── components/           # React コンポーネント
│   │   │   ├── Auth/            # 認証関連
│   │   │   ├── CardManagement/  # カード管理
│   │   │   ├── DeckManagement/  # デッキ管理
│   │   │   ├── GameBoard.tsx    # ゲームボード
│   │   │   └── ...
│   │   ├── gen/                  # 自動生成コード（TypeScript）
│   │   ├── hooks/                # カスタムフック
│   │   ├── lib/                  # ライブラリ（API クライアント等）
│   │   ├── pages/                # ページコンポーネント
│   │   ├── types/                # 型定義
│   │   └── utils/                # ユーティリティ
│   ├── buf.yaml                  # Buf 設定
│   └── package.json
├── mysql/volumes/                 # MySQL データディレクトリ
├── docker-compose.yml             # Docker Compose 設定
├── Makefile                       # タスクランナー
└── README.md

```

詳細なアーキテクチャについては [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) を参照してください。

## API 仕様

Protocol Buffers で定義された API は以下の4つのサービスで構成されています：

### 1. AuthService (認証)
- `Register`: 新規ユーザー登録
- `Login`: ログイン（JWT 発行）

### 2. CardManagementService (カード管理)
- `GetAllCards`: 全カード取得
- `CreateCard`: カード作成（管理者）

### 3. DeckManagementService (デッキ管理)
- `CreateDeck`: デッキ作成
- `GetDecks`: デッキ一覧取得
- `GetDeck`: デッキ詳細取得
- `UpdateDeck`: デッキ更新
- `DeleteDeck`: デッキ削除

### 4. GameService (ゲーム)
- `CreateGame`: ゲーム作成
- `JoinGame`: ゲーム参加
- `PerformMulligan`: マリガン実行
- `PlayCard`: カードプレイ
- `Attack`: 攻撃
- `EndTurn`: ターン終了
- `GetGameState`: ゲーム状態取得
- `StreamGameEvents`: ゲームイベントストリーム（Server Streaming）

## 開発コマンド

```bash
# ヘルプを表示
make help

# Protocol Buffers からコードを生成
make proto

# 依存関係のインストール
make deps

# Protocol Buffers ツールのインストール
make install-tools

# ビルド
make build-connect
make build              # 上記のエイリアス

# サーバー起動
make run-connect

# テスト実行
make test

# クリーンアップ
make clean
```

### フロントエンド用コマンド

```bash
cd frontend

# 開発サーバー起動
pnpm dev

# ビルド
pnpm build

# Protocol Buffers からコード生成
pnpm proto:generate

# Lint
pnpm lint
pnpm lint:fix

# Format
pnpm format

# Lint + Format 統合チェック
pnpm check
pnpm check:fix
```

## データベース管理

```bash
# MySQL 起動
docker-compose up -d

# MySQL 停止
docker-compose down

# ログ確認
docker-compose logs -f db

# MySQL コンテナに接続
docker-compose exec db mysql -u root -p
```

## トラブルシューティング

### バックエンドが起動しない
- `.env` ファイルが存在し、正しく設定されているか確認
- MySQL が起動しているか確認: `docker-compose ps`
- ポート 8080 が使用されていないか確認: `lsof -i :8080`

### フロントエンドが起動しない
- `pnpm install` を実行したか確認
- `pnpm proto:generate` を実行したか確認
- ポート 5173 が使用されていないか確認

### Protocol Buffers の生成に失敗
- `protoc` がインストールされているか確認: `protoc --version`
- `make install-tools` を実行
- Go の PATH が通っているか確認

### 認証エラー
- JWT トークンが正しく設定されているか確認
- `.env` の `JWT_SECRET` が設定されているか確認
- ブラウザの LocalStorage を確認・クリア

## ライセンス

MIT License


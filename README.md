# Card Game

オンライン対戦型カードゲームプロジェクト

## 概要

このプロジェクトは、Web上で動作する対戦型カードゲームです。Go言語とConnect-RPCで構築されたバックエンドと、React + TypeScriptで構築されたフロントエンドから構成されています。

## 技術スタック

### バックエンド
- **言語**: Go 1.25.4
- **通信**: Connect-RPC (gRPC/HTTP/JSON)
- **データベース**: MySQL 8.4
- **ORM**: GORM
- **認証**: JWT (golang-jwt)
- **アーキテクチャ**: クリーンアーキテクチャ

### フロントエンド
- **フレームワーク**: React 18
- **言語**: TypeScript
- **ビルドツール**: Vite
- **API通信**: Connect-Web
- **コード品質**: Biome

## アーキテクチャ

### クリーンアーキテクチャの採用

プロジェクトはクリーンアーキテクチャに基づいて設計されており、以下のレイヤーに分離されています：

```
internal/
├── core/              # ビジネスロジックの中核（最も内側）
│   ├── entity/        # エンティティ: ビジネスルールを持つドメインオブジェクト
│   ├── usecase/       # ユースケース: アプリケーション固有のビジネスルール
│   └── port/          # ポート: 外部層とのインターフェース定義
├── application/       # アプリケーション層
│   └── service/       # アプリケーションサービス（ユースケースの調整）
├── adapter/           # アダプター層（外部とのやり取り）
│   ├── connect/       # Connect-RPCハンドラー
│   ├── converter/     # データ変換
│   └── jwt/           # JWT関連
└── infrastructure/    # インフラストラクチャ層（最も外側）
    ├── repository/    # データベースアクセスの実装
    ├── persistence/   # データベース接続管理
    ├── auth/          # 認証実装
    └── logger/        # ロギング実装
```

#### レイヤーの役割

- **Core**: 依存関係を持たない純粋なビジネスロジック
- **Application**: コアのユースケースを組み合わせてアプリケーション機能を提供
- **Adapter**: 外部インターフェース（API、CLI等）とコアを繋ぐ
- **Infrastructure**: 外部システム（DB、認証等）の具体的な実装

### 依存関係の方向

依存関係は常に外側から内側に向かいます：

```
Infrastructure → Adapter → Application → Core
```

コア層は他のどの層にも依存せず、Portインターフェースを通じて外部とやり取りします。

## セットアップ

### 前提条件

- Go 1.25.4以上
- Node.js (LTS版推奨)
- pnpm
- Docker & Docker Compose
- Protocol Buffers コンパイラ (protoc)

### 環境変数設定

```bash
cp example.env .env
# .envファイルを編集して必要な値を設定
```

### データベース起動

```bash
docker-compose up -d
```

### バックエンド起動

```bash
# 依存パッケージのインストール
make deps

# Protoファイルからコードを生成
make proto

# ビルド
make build

# サーバー起動
make run-connect
```

### フロントエンド起動

```bash
cd frontend

# 依存パッケージのインストール
pnpm install

# Protoファイルからコードを生成
pnpm proto:generate

# 開発サーバー起動
pnpm dev
```

## 開発コマンド

### バックエンド

```bash
make help           # 使用可能なコマンド一覧を表示
make proto          # Protoファイルからコードを生成
make build          # アプリケーションをビルド
make run-connect    # サーバーを起動
```

### フロントエンド

```bash
pnpm dev            # 開発サーバー起動
pnpm build          # 本番用ビルド
pnpm lint           # コードチェック
pnpm format         # コードフォーマット
```

## セキュリティ機能

### 実装済み機能

1. **認証・認可**
   - JWT認証（HMAC-SHA256）
   - bcryptパスワードハッシュ化（コスト12-14）
   - トークン有効期限管理

2. **レート制限**
   - アクション毎の制限設定
   - スライディングウィンドウ方式
   - 自動クリーンアップ

3. **不正検出**
   - ゲーム状態検証（ハッシュ比較）
   - チート検出（不可能なタイミング検出）
   - 疑わしいアクションのスコアリング

4. **入力検証**
   - SQLインジェクション対策
   - XSS対策
   - 入力サニタイゼーション

5. **セキュリティヘッダー**
   - X-Content-Type-Options
   - X-Frame-Options
   - X-XSS-Protection
   - Referrer-Policy

### セキュリティ設定

詳細は以下のドキュメントを参照してください：

- [SECURITY.md](docs/SECURITY.md) - セキュリティガイド
- [.env.example](.env.example) - 環境変数設定例
- `scripts/generate-secrets.sh` - ランダムキー生成スクリプト

### セキュリティのベストプラクティス

⚠️ **本番環境デプロイ前の必須事項**:

1. `.env.example`をコピーして`.env`を作成
2. すべてのシークレット値を強力なランダム値に変更
3. JWT_SECRETは32文字以上のランダム文字列
4. データベースパスワードは強力なパスワードに変更
5. `.env`ファイルをGitにコミットしない

```bash
# セキュアなランダムキーを生成
./scripts/generate-secrets.sh
```

## API仕様

API仕様はProtocol Buffersで定義されています：

- `api/proto/auth.proto` - 認証関連
- `api/proto/card_management.proto` - カード管理
- `api/proto/game.proto` - ゲームロジック
- `api/proto/common.proto` - 共通定義

## ゲームの仕様

### ゲームの流れ

1. **デッキ選択**: プレイヤーは自分のデッキを選択
2. **マリガン**: 初期手札を引き直すか選択
3. **ターン制バトル**: 交互にターンを行い対戦
4. **勝利条件**: 相手プレイヤーのHPを0にする

### カードの種類

- **ユニットカード**: フィールドに配置して戦闘を行う
- **エンチャント**: ユニットに付与して能力を強化

詳細な仕様については、ゲーム内のルールおよびコード内のドキュメントを参照してください。

## プロジェクト構造の詳細

### バックエンド主要ディレクトリ

- `cmd/connect-server/` - サーバーのエントリーポイント
- `api/proto/` - Protocol Buffers定義ファイル
- `internal/` - アプリケーションコア（クリーンアーキテクチャ）

### フロントエンド主要ディレクトリ

- `src/components/` - Reactコンポーネント
- `src/services/` - APIクライアント
- `src/hooks/` - カスタムフック
- `src/domain/` - ドメインモデル
- `src/gen/` - Protoから自動生成されたコード

## ライセンス

未定

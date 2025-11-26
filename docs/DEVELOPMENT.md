# 開発ガイド

## 開発環境のセットアップ

### 必要なツール

- **Go**: 1.25.4以上
- **Node.js**: LTS版推奨
- **pnpm**: パッケージマネージャー
- **Docker Desktop**: データベース環境
- **Protocol Buffers コンパイラ**: `protoc`

### 初回セットアップ

```bash
# 1. リポジトリのクローン
git clone <repository-url>
cd card_game

# 2. 環境変数の設定
cp example.env .env
# .envを編集して必要な値を設定

# 3. Goツールのインストール
make install-tools

# 4. 依存パッケージのインストール
make deps

# 5. データベースの起動
docker-compose up -d

# 6. Protoファイルからコードを生成
make proto

# 7. フロントエンドのセットアップ
cd frontend
pnpm install
pnpm proto:generate
cd ..
```

## 開発フロー

### バックエンド開発

#### 1. Protoファイルの更新

APIを追加・変更する場合、まずProtoファイルを更新します。

```bash
# api/proto/ 内の.protoファイルを編集

# コード生成
make proto
```

#### 2. コアロジックの実装

クリーンアーキテクチャに従って実装：

1. **Entity定義** (`internal/core/entity/`)
   - ドメインモデルを定義
   - ビジネスルールをメソッドとして実装

2. **Port定義** (`internal/core/port/`)
   - 必要なインターフェースを定義
   - 例: `type CardRepository interface { ... }`

3. **UseCase実装** (`internal/core/usecase/`)
   - ビジネスロジックを実装
   - Portインターフェースを使用

#### 3. インフラストラクチャ実装

```bash
# Repository実装
internal/infrastructure/repository/

# その他のインフラストラクチャ
internal/infrastructure/persistence/
internal/infrastructure/auth/
```

#### 4. アプリケーションサービス実装

```bash
# internal/application/service/
# ユースケースを組み合わせた機能を実装
```

#### 5. ハンドラー実装

```bash
# internal/adapter/connect/handler/
# APIエンドポイントを実装
```

#### 6. サーバー起動

```bash
make run-connect
```

### フロントエンド開発

#### 1. Protoから型生成

バックエンドでProtoファイルを更新した後：

```bash
cd frontend
pnpm proto:generate
```

#### 2. コンポーネント開発

```bash
# src/components/ にコンポーネントを作成
# src/hooks/ にカスタムフックを作成
# src/services/ にAPIクライアントを作成
```

#### 3. 開発サーバー起動

```bash
pnpm dev
```

#### 4. コード品質チェック

```bash
# Lintチェック
pnpm lint

# 自動修正
pnpm lint:fix

# フォーマット
pnpm format

# 全チェック
pnpm check
```

## デバッグ

### バックエンドのデバッグ

#### ログ出力

標準のロガーインターフェースを使用：

```go
logger.Info("メッセージ")
logger.Error("エラーメッセージ")
```

#### データベースの確認

```bash
# MySQLに接続
docker exec -it card_game-mysql mysql -u root -p

# データベース選択
USE card_game;

# テーブル確認
SHOW TABLES;
SELECT * FROM users;
```

### フロントエンドのデバッグ

#### ブラウザの開発者ツール

- Console: ログ出力を確認
- Network: API通信を確認
- React DevTools: コンポーネント状態を確認

## テスト

### バックエンドテスト

```bash
# 全テスト実行
make test

# カバレッジ付き
make test-coverage
```

### フロントエンドテスト

```bash
cd frontend
pnpm test
```

## ビルドとデプロイ

### バックエンドビルド

```bash
make build
# 生成物: bin/connect-server
```

### フロントエンドビルド

```bash
cd frontend
pnpm build
# 生成物: dist/
```

## トラブルシューティング

### よくある問題

#### データベース接続エラー

```bash
# コンテナの状態を確認
docker-compose ps

# ログを確認
docker-compose logs db

# 再起動
docker-compose restart db
```

#### Protoコード生成エラー

```bash
# ツールを再インストール
make install-tools

# 再生成
make proto
```

#### フロントエンドのビルドエラー

```bash
cd frontend

# node_modulesを削除して再インストール
rm -rf node_modules pnpm-lock.yaml
pnpm install
```

## コーディング規約

### Go

- 標準の `gofmt` でフォーマット
- `golint` でリンティング
- エラーハンドリングを適切に実装
- インターフェースは小さく保つ

### TypeScript/React

- Biomeの設定に従う
- 関数コンポーネントを使用
- カスタムフックで状態管理ロジックを分離
- PropTypesではなくTypeScriptの型を使用

## Git運用

### ブランチ戦略

```
main: 本番環境のコード
develop: 開発中のコード
feature/*: 機能開発ブランチ
fix/*: バグ修正ブランチ
```

### コミットメッセージ

```
feat: 新機能
fix: バグ修正
docs: ドキュメント変更
refactor: リファクタリング
test: テスト追加・変更
chore: その他の変更
```

## 参考資料

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Connect-RPC Documentation](https://connectrpc.com/)
- [Protocol Buffers](https://protobuf.dev/)
- [React Documentation](https://react.dev/)

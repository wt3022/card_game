# Card Game

デジタルカードバトルゲーム

## 概要

Go + React で実装されたターン制のカードゲームです。Connect-RPC を使用してバックエンドとフロントエンドが通信します。

## 技術スタック

- **バックエンド**: Go 1.25.4 + Connect-RPC
- **フロントエンド**: React 18 + TypeScript + Vite
- **通信**: Protocol Buffers (protobuf)
- **アーキテクチャ**: クリーンアーキテクチャ

## セットアップ

### 必要なツール

- Go 1.25以上
- Node.js 18以上
- pnpm
- protoc (Protocol Buffers コンパイラ)

### インストール

```bash
# バックエンドの依存関係
make deps

# protoツールのインストール
make install-tools

# フロントエンドの依存関係
cd frontend
pnpm install
```

## 実行方法

### 開発環境

```bash
# バックエンドサーバーを起動
make run-connect

# フロントエンド開発サーバーを起動（別ターミナル）
cd frontend
pnpm dev
```

バックエンド: http://localhost:8080
フロントエンド: http://localhost:3000

### ビルド

```bash
# バックエンド
make build-connect

# フロントエンド
cd frontend
pnpm build
```

## ゲームの基本ルール

- 各プレイヤーは初期HP 20、初期マナ 1でスタート
- 初期手札は4枚
- ターンごとにマナが回復・増加
- ユニットカードを召喚し、相手プレイヤーや相手ユニットを攻撃
- スペルカードで効果を発動
- 相手のHPを0にすると勝利

## プロジェクト構成

詳細は [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) を参照してください。

## 開発コマンド

```bash
make help          # 利用可能なコマンドを表示
make proto         # protoファイルからコードを生成
make test          # テストを実行
make clean         # 生成ファイルを削除
```

## ライセンス

MIT


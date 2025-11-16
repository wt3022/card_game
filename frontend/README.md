# Card Game Frontend

Reactを使用したカードゲームのフロントエンドクライアント

## 特徴

- 🎮 **リアルタイムマッチング** - サーバーストリーミングで対戦相手を自動マッチング
- 🎴 **手札管理** - カードの詳細情報を表示し、直感的に使用可能
- ⚔️ **ターン制バトル** - プレイヤー情報、フィールド、手札を1画面で管理
- 📱 **レスポンシブUI** - スクロール不要の最適化されたレイアウト

## セットアップ

### 1. 依存関係のインストール

```bash
npm install
```

### 2. 型定義の生成

protoファイルから型定義を自動生成します：

```bash
npm run proto:generate
```

これにより、`src/gen`ディレクトリに以下のファイルが生成されます：
- `common_pb.ts` - 共通の型定義
- `game_pb.ts` - ゲーム関連の型定義
- `game_connect.ts` - Connect-Web用のサービスクライアント

### 3. 開発サーバーの起動

```bash
npm run dev
```

アプリケーションは http://localhost:3000 で起動します。

バックエンドサーバーが http://localhost:8080 で動作している必要があります。

## ビルド

本番用にビルド：

```bash
npm run build
```

ビルドされたファイルは`dist`ディレクトリに出力されます。

## プレビュー

ビルドしたアプリケーションをプレビュー：

```bash
npm run preview
```

## 技術スタック

- **React 18** - UIフレームワーク
- **TypeScript** - 型安全性
- **Vite** - ビルドツール
- **Connect-Web** - gRPC-Webクライアント
- **Buf** - Protobufコード生成

## プロジェクト構造

```
frontend/
├── src/
│   ├── components/       # Reactコンポーネント
│   │   ├── GameSetup.tsx      # ゲーム開始画面
│   │   ├── GameBoard.tsx      # ゲームボード
│   │   ├── PlayerInfo.tsx     # プレイヤー情報
│   │   ├── UnitCard.tsx       # ユニットカード
│   │   └── HandCard.tsx       # 手札カード
│   ├── lib/              # ユーティリティ
│   │   └── api-client.ts      # APIクライアント
│   ├── gen/              # 自動生成された型定義（gitignore）
│   ├── App.tsx           # メインアプリケーション
│   ├── main.tsx          # エントリーポイント
│   └── index.css         # グローバルスタイル
├── buf.gen.yaml          # Buf設定ファイル
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## ゲームの流れ

### 1. マッチング
1. プレイヤー名を入力
2. 「マッチング開始」をクリック
3. 他のプレイヤーがマッチングに参加するまで待機
4. マッチング成功！→ 自動的にゲーム開始

### 2. ゲームプレイ
- **手札からカード使用** - カードをクリック→「使用」ボタン
- **ユニットで攻撃** - 自分のユニットをクリック→相手のユニットまたはプレイヤーをクリック
- **ターン終了** - 「ターン終了」ボタンをクリック

## API連携

バックエンドとの通信には、Connect-Webを使用しています。
`src/lib/api-client.ts`でクライアントを設定し、以下のAPIを利用できます：

- `joinMatchmaking` - マッチングに参加（サーバーストリーミング）⭐新機能
- `getGameState` - ゲーム状態の取得
- `playCard` - カードのプレイ
- `executeAttack` - 攻撃の実行
- `endTurn` - ターンの終了

## 開発メモ

### 型定義の更新

protoファイルを更新した場合は、型定義を再生成してください：

```bash
npm run proto:generate
```

### バックエンドとの接続

開発時は、Viteのプロキシ設定により`/cardgame.v1.GameService`へのリクエストが
バックエンドサーバー（http://localhost:8080）に転送されます。

本番環境では、適切なCORS設定とURLを`src/lib/api-client.ts`で設定してください。


# アーキテクチャ

## クリーンアーキテクチャ

本プロジェクトはクリーンアーキテクチャを採用し、依存関係の方向を内側（ドメイン）に向けることで、ビジネスロジックを独立させています。

## ディレクトリ構成

```
card_game/
├── api/                           # API定義
│   ├── proto/                     # Protocol Buffers定義
│   └── gen/                       # 自動生成されたコード
├── cmd/                           # エントリポイント
│   └── connect-server/            # Connect-RPCサーバー
├── internal/
│   ├── core/                      # コア層（ドメイン層）
│   │   ├── entity/                # エンティティ
│   │   ├── usecase/               # ユースケース（ビジネスロジック）
│   │   └── port/                  # インターフェース定義
│   ├── application/               # アプリケーション層
│   │   └── service/               # アプリケーションサービス
│   ├── adapter/                   # アダプター層
│   │   ├── connect/               # Connect-RPCハンドラー
│   │   └── converter/             # DTO変換
│   └── infrastructure/            # インフラ層
│       └── event/                 # イベント管理
└── frontend/                      # フロントエンド
    └── src/
        ├── components/            # Reactコンポーネント
        ├── gen/                   # 自動生成されたprotobufコード
        └── lib/                   # ユーティリティ
```

## レイヤー構成

### 1. Core層（内側）

**entity/** - ドメインエンティティ
- `card.go`: カードの定義（ユニット/スペル/リーダー）
- `player.go`: プレイヤーの状態管理
- `unit.go`: フィールド上のユニット
- `effect.go`: カード効果の定義
- `trait.go`: カードの特性（Taunt, Rush等）
- `event.go`: ゲームイベントの記録

**usecase/** - ビジネスロジック
- `engine.go`: ゲームエンジンのコア
- `card_play.go`: カードプレイの処理
- `combat/`: 戦闘システム
- `effect/`: 効果処理システム
  - `processor.go`: 効果の実行エンジン
  - `atomic/`: 原子効果（ダメージ、回復、ドロー等）
- `game/`: ゲーム進行管理
  - `state.go`: ゲーム状態管理
  - `phase.go`: フェーズ管理
  - `turn.go`: ターン処理

**port/** - インターフェース定義
- 外部システムとの契約を定義
- Loggerやイベントバスのインターフェース

### 2. Application層

**service/** - アプリケーションサービス
- `game_service.go`: ゲームセッション管理
- マッチメイキング機能
- ゲーム状態の永続化（将来拡張）

### 3. Adapter層（外側）

**connect/handler/** - Connect-RPCハンドラー
- protoで定義されたAPIのハンドラー実装
- DTO変換とバリデーション

**converter/** - データ変換
- ドメインモデル ↔ protobuf DTO の変換

### 4. Infrastructure層（最外層）

**event/** - イベント管理
- ゲームイベントの配信
- リアルタイム通信のサポート

## 依存関係の方向

```
Infrastructure → Adapter → Application → Core
                                           ↑
                                        （依存しない）
```

Core層は他のどの層にも依存せず、純粋なビジネスロジックを保持します。

## 効果システム

カードの効果は**原子効果**の組み合わせで表現されます：

- **原子効果**: ダメージ、回復、ドロー、バフ/デバフ等の基本操作
- **演算子**: 効果を組み合わせる（AND, OR, THEN等）
- **条件**: 効果発動の条件判定
- **ターゲット**: 効果の対象選択

例: 「敵全体に2ダメージ + 味方全体を2回復」
```
[DEAL_SPLASH(2, enemy)] THEN [RESTORE_HP(2, ally)]
```

## 通信プロトコル

- **Connect-RPC**: HTTP/JSON または gRPC で通信
- **Protocol Buffers**: 型安全なAPI定義
- **ストリーミング**: リアルタイムイベント配信に対応

## フロントエンド

- **React + TypeScript**: コンポーネントベースのUI
- **Vite**: 高速な開発サーバー
- **@connectrpc/connect-web**: Connect-RPCクライアント
- **Protocol Buffers**: バックエンドと型定義を共有


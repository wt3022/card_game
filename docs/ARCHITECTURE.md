# アーキテクチャ設計書

## クリーンアーキテクチャ

本プロジェクトはクリーンアーキテクチャ（Clean Architecture）を採用し、依存関係の方向を内側（ドメイン）に向けることで、ビジネスロジックを外部の詳細実装から独立させています。

### クリーンアーキテクチャの原則

1. **依存関係逆転の原則**: 外側の層が内側の層に依存し、内側は外側を知らない
2. **関心の分離**: 各層は明確な責務を持ち、単一責任の原則に従う
3. **テスタビリティ**: ドメインロジックを独立してテスト可能
4. **フレームワーク非依存**: ビジネスロジックは特定のフレームワークに依存しない

## ディレクトリ構成

```
card_game/
├── api/                           # API定義
│   ├── proto/                     # Protocol Buffers定義
│   │   ├── auth.proto            # 認証API
│   │   ├── card_management.proto # カード管理API
│   │   ├── game.proto            # ゲームAPI
│   │   └── common.proto          # 共通型定義
│   └── gen/proto/cardgame/v1/    # 自動生成されたコード
│       ├── *_pb.go               # Protocol Buffers Go コード
│       └── *_connect.go          # Connect-RPC サーバー・クライアントコード
│
├── cmd/                           # エントリポイント
│   └── connect-server/
│       └── main.go               # Connect-RPCサーバーのメイン関数
│
├── internal/                      # 内部パッケージ
│   ├── core/                      # コア層（ドメイン層）- 最も内側
│   │   ├── entity/                # エンティティ（ドメインモデル）
│   │   │   ├── card.go           # カード（Unit/Spell/Leader）
│   │   │   ├── player.go         # プレイヤー
│   │   │   ├── unit.go           # フィールド上のユニット
│   │   │   ├── effect.go         # カード効果
│   │   │   ├── trait.go          # カード特性（Taunt, Rush等）
│   │   │   ├── event.go          # ゲームイベント
│   │   │   ├── deck.go           # デッキ
│   │   │   └── user.go           # ユーザー
│   │   │
│   │   ├── usecase/               # ユースケース（ビジネスロジック）
│   │   │   ├── engine.go         # ゲームエンジンコア
│   │   │   ├── card_play.go      # カードプレイ処理
│   │   │   ├── combat/           # 戦闘システム
│   │   │   │   ├── attack.go    # 攻撃処理
│   │   │   │   └── damage.go    # ダメージ計算
│   │   │   ├── effect/           # 効果処理システム
│   │   │   │   ├── processor.go # 効果実行エンジン
│   │   │   │   └── atomic/      # 原子効果（ダメージ、回復、ドロー等）
│   │   │   └── game/             # ゲーム進行管理
│   │   │       ├── state.go     # ゲーム状態管理
│   │   │       ├── phase.go     # フェーズ管理
│   │   │       └── turn.go      # ターン処理
│   │   │
│   │   └── port/                  # インターフェース定義（ポート）
│   │       ├── logger.go         # ロガーインターフェース
│   │       ├── event_bus.go      # イベントバスインターフェース
│   │       └── repository.go     # リポジトリインターフェース
│   │
│   ├── application/               # アプリケーション層
│   │   └── service/               # アプリケーションサービス
│   │       ├── auth_service.go   # 認証サービス（ユーザー登録・ログイン）
│   │       ├── card_service.go   # カード管理サービス
│   │       ├── deck_service.go   # デッキ管理サービス
│   │       ├── game_service.go   # ゲームセッション管理
│   │       └── user_init.go      # ユーザー初期化
│   │
│   ├── adapter/                   # アダプター層（外部インターフェース）
│   │   ├── connect/               # Connect-RPC関連
│   │   │   ├── handler/          # APIハンドラー
│   │   │   │   ├── auth_handler.go        # 認証API
│   │   │   │   ├── card_handler.go        # カード管理API
│   │   │   │   ├── deck_handler.go        # デッキ管理API
│   │   │   │   └── game_handler.go        # ゲームAPI
│   │   │   └── interceptor/      # インターセプター
│   │   │       └── auth_interceptor.go    # JWT認証
│   │   ├── converter/             # DTO変換
│   │   │   ├── from_proto.go     # proto → entity
│   │   │   └── to_proto.go       # entity → proto
│   │   └── jwt/                   # JWT処理
│   │       └── jwt.go
│   │
│   ├── infrastructure/            # インフラ層（最外層）
│   │   ├── auth/                 # 認証基盤
│   │   │   └── password.go      # パスワードハッシュ化
│   │   ├── event/                # イベント管理
│   │   │   └── bus.go           # イベントバス実装
│   │   ├── logger/               # ロガー実装
│   │   │   └── console.go       # コンソールロガー
│   │   ├── persistence/          # データベース接続
│   │   │   ├── db.go            # DB接続管理
│   │   │   └── migration.go     # マイグレーション
│   │   └── repository/           # データアクセス実装
│   │       ├── user_repository.go
│   │       ├── card_repository.go
│   │       └── deck_repository.go
│   │
│   └── fixture/                   # テストフィクスチャ・初期データ
│       └── deck/                 # デッキ定義
│
├── frontend/                      # フロントエンド
│   ├── src/
│   │   ├── components/            # Reactコンポーネント
│   │   │   ├── Auth/             # 認証関連コンポーネント
│   │   │   ├── CardManagement/   # カード管理
│   │   │   ├── DeckManagement/   # デッキ管理
│   │   │   ├── GameBoard.tsx     # ゲームボード
│   │   │   ├── GameSetup.tsx     # ゲーム開始画面
│   │   │   ├── MulliganModal.tsx # マリガン画面
│   │   │   └── ...
│   │   ├── gen/                   # 自動生成されたコード
│   │   │   ├── *_pb.ts           # Protocol Buffers TypeScript
│   │   │   └── *_connect.ts      # Connect-RPC クライアント
│   │   ├── hooks/                 # カスタムフック
│   │   │   ├── useGameState.ts   # ゲーム状態管理
│   │   │   ├── useGameActions.ts # ゲームアクション
│   │   │   ├── useMulligan.ts    # マリガン処理
│   │   │   └── ...
│   │   ├── lib/                   # ライブラリ
│   │   │   ├── api-client.ts     # API クライアント
│   │   │   └── auth.ts           # 認証ヘルパー
│   │   ├── pages/                 # ページコンポーネント
│   │   │   ├── Game.tsx          # ゲームページ
│   │   │   └── Admin.tsx         # 管理画面
│   │   ├── types/                 # 型定義
│   │   └── utils/                 # ユーティリティ
│   ├── buf.yaml                   # Buf設定（protobuf生成）
│   └── package.json
│
├── mysql/volumes/                 # MySQLデータディレクトリ
├── docker-compose.yml             # Docker Compose設定
├── Makefile                       # タスクランナー
└── go.mod                         # Go モジュール定義
```

## レイヤー構成

### 1. Core層（最内層）- ドメインロジック

**依存**: なし（他のどの層にも依存しない）

#### entity/ - ドメインエンティティ

ビジネスの中核となるドメインモデルを定義します。

- `card.go`: カードの定義
  - カードタイプ（Unit / Spell / Leader）
  - コスト、攻撃力、体力、効果
- `player.go`: プレイヤーの状態管理
  - HP、マナ、デッキ、手札、フィールド
- `unit.go`: フィールド上のユニット
  - 攻撃力、体力、状態（召喚酔い、攻撃済み等）
- `effect.go`: カード効果の定義
  - 効果タイプ、ターゲット、条件
- `trait.go`: カードの特性
  - Taunt（挑発）、Rush（速攻）、Divine Shield（聖なる盾）等
- `event.go`: ゲームイベントの記録
  - カードプレイ、攻撃、ダメージ、死亡等のイベント
- `deck.go`: デッキ構成
- `user.go`: ユーザー情報

#### usecase/ - ユースケース（ビジネスロジック）

ゲームのルールとロジックを実装します。

- `engine.go`: ゲームエンジンのコア
  - ゲーム全体の制御フロー
  - ルール検証
- `card_play.go`: カードプレイの処理
  - マナコストチェック
  - カード効果の発動
- `combat/`: 戦闘システム
  - `attack.go`: 攻撃処理（ターゲット選択、Tauntチェック）
  - `damage.go`: ダメージ計算と適用
- `effect/`: 効果処理システム
  - `processor.go`: 効果の実行エンジン
  - `atomic/`: 原子効果（ダメージ、回復、ドロー、バフ/デバフ等）
    - ダメージ、回復、カードドロー等の基本操作
    - 効果の連鎖処理
- `game/`: ゲーム進行管理
  - `state.go`: ゲーム状態管理（誰のターンか、フェーズ、勝敗判定）
  - `phase.go`: フェーズ管理（ドロー、メイン、バトル、エンド）
  - `turn.go`: ターン処理（マナ回復、ドロー実行）

#### port/ - インターフェース定義（ポート）

外部システムとの契約を定義します（依存性逆転の原則）。

- `logger.go`: ロガーインターフェース
- `event_bus.go`: イベントバスインターフェース
- `repository.go`: リポジトリインターフェース（データ永続化）

### 2. Application層 - アプリケーションサービス

**依存**: Core層のみ

ユースケースを組み合わせてアプリケーション固有の機能を実現します。

#### service/ - アプリケーションサービス

- `auth_service.go`: 認証サービス
  - ユーザー登録、ログイン
  - JWTトークン発行（JWT自体の生成はadapter層）
- `card_service.go`: カード管理サービス
  - カードの作成、取得
  - カードマスターデータ管理
- `deck_service.go`: デッキ管理サービス
  - デッキ作成、更新、削除
  - デッキバリデーション（30枚、カード重複制限等）
- `game_service.go`: ゲームセッション管理
  - ゲーム作成、参加
  - ゲーム状態の管理
  - エンジンとの連携
- `user_init.go`: ユーザー初期化処理

### 3. Adapter層 - 外部インターフェース

**依存**: Core層、Application層

外部からのリクエストを受け取り、内部のドメインロジックに変換します。

#### connect/handler/ - Connect-RPCハンドラー

Protocol Buffersで定義されたAPIのハンドラー実装。

- `auth_handler.go`: 認証API
  - `Register`: ユーザー登録
  - `Login`: ログイン
- `card_handler.go`: カード管理API
  - `GetAllCards`: 全カード取得
  - `CreateCard`: カード作成（管理者）
- `deck_handler.go`: デッキ管理API
  - `CreateDeck`, `GetDecks`, `GetDeck`, `UpdateDeck`, `DeleteDeck`
- `game_handler.go`: ゲームAPI
  - `CreateGame`, `JoinGame`, `PerformMulligan`
  - `PlayCard`, `Attack`, `EndTurn`
  - `GetGameState`, `StreamGameEvents`（Server Streaming）

#### connect/interceptor/ - インターセプター

- `auth_interceptor.go`: JWT認証インターセプター
  - リクエストヘッダーからトークン抽出
  - トークン検証
  - ユーザーIDをコンテキストに設定

#### converter/ - データ変換

Protocol Buffers DTO とドメインエンティティ間の変換。

- `from_proto.go`: proto → entity（リクエスト変換）
- `to_proto.go`: entity → proto（レスポンス変換）

#### jwt/ - JWT処理

- `jwt.go`: JWTトークンの生成・検証

### 4. Infrastructure層（最外層）- 実装詳細

**依存**: すべての層（ただしCore層のポートインターフェースを実装）

フレームワークやライブラリに依存する実装を提供します。

#### auth/ - 認証基盤

- `password.go`: パスワードハッシュ化（bcrypt）

#### event/ - イベント管理

- `bus.go`: イベントバス実装
  - ゲームイベントの配信
  - Server Streamingでクライアントに送信

#### logger/ - ロガー実装

- `console.go`: コンソールロガー
  - Core層のLoggerインターフェースを実装

#### persistence/ - データベース接続

- `db.go`: MySQL接続管理
  - GORM初期化
- `migration.go`: マイグレーション処理

#### repository/ - データアクセス実装

Core層のRepositoryインターフェースを実装。

- `user_repository.go`: ユーザーデータアクセス
- `card_repository.go`: カードデータアクセス
- `deck_repository.go`: デッキデータアクセス

## 依存関係の方向

```
Infrastructure → Adapter → Application → Core
                 ↓           ↓            ↑
              (実装)      (使用)    (インターフェース定義)
```

- **Core層**: 他のどの層にも依存しない（依存関係の中心）
- **Application層**: Core層のみに依存
- **Adapter層**: Core層とApplication層に依存
- **Infrastructure層**: すべての層に依存可能（ただしCore層のインターフェースを実装）

この設計により、以下のメリットが得られます：

1. **テスタビリティ**: Core層を独立してテスト可能
2. **保守性**: ビジネスロジックとインフラの分離
3. **拡張性**: 新しいアダプターやインフラを容易に追加可能
4. **フレームワーク独立性**: Core層はフレームワークに依存しない

## 効果システム

カードの効果は**原子効果**（Atomic Effect）の組み合わせで表現されます。

### 設計思想

効果を小さな単位（原子効果）に分解し、それらを組み合わせることで複雑な効果を実現します。

### 構成要素

#### 1. 原子効果（Atomic Effect）

基本的な操作単位。

- **ダメージ**: `DEAL_DAMAGE(target, amount)`
- **範囲ダメージ**: `DEAL_SPLASH(targets, amount)`
- **回復**: `RESTORE_HP(target, amount)`
- **カードドロー**: `DRAW_CARD(player, count)`
- **バフ**: `BUFF_ATTACK(target, amount)`, `BUFF_HP(target, amount)`
- **デバフ**: `DEBUFF_ATTACK(target, amount)`, `DEBUFF_HP(target, amount)`
- **特性付与**: `GRANT_TRAIT(target, trait)`（例: Taunt, Rush）
- **召喚**: `SUMMON(card, position)`

#### 2. 演算子（Operator）

効果を組み合わせる。

- **AND**: 複数の効果を同時実行
- **OR**: 条件によって効果を選択
- **THEN**: 効果を順次実行

#### 3. 条件（Condition）

効果発動の条件判定。

- ターゲットの状態（HP、特性等）
- ゲームの状態（ターン数、フィールドの状況等）

#### 4. ターゲット（Target）

効果の対象選択。

- 単体ターゲット: 特定のユニットまたはプレイヤー
- 範囲ターゲット: すべての敵ユニット、すべての味方ユニット等
- ランダムターゲット: ランダムな敵ユニット等

### 効果の例

#### 例1: 「敵全体に2ダメージ + 味方全体を2回復」

```
[DEAL_SPLASH(enemy_units, 2)] THEN [RESTORE_HP(ally_units, 2)]
```

#### 例2: 「ランダムな敵ユニット1体に5ダメージ、そのユニットを破壊した場合、カードを1枚ドロー」

```
[DEAL_DAMAGE(random_enemy, 5)] THEN
IF (target_destroyed) THEN [DRAW_CARD(self, 1)]
```

#### 例3: 「味方ユニット全体に+2/+2のバフを付与し、Rushを与える」

```
[BUFF_ATTACK(ally_units, 2)] AND [BUFF_HP(ally_units, 2)] AND [GRANT_TRAIT(ally_units, Rush)]
```

### 効果処理の流れ

1. **効果の定義**: カードに効果が紐づけられる（`entity/effect.go`）
2. **効果の発動**: カードがプレイされると、効果が発動される
3. **ターゲット選択**: ユーザーまたは自動でターゲットを選択
4. **効果の実行**: `usecase/effect/processor.go` が原子効果を順次実行
5. **イベント記録**: 各効果の結果をイベントとして記録
6. **クライアント配信**: イベントをServer Streamingでクライアントに送信

## 通信プロトコル

### Connect-RPC

HTTP/1.1、HTTP/2、gRPCに対応した RPC フレームワーク。

- **Unary RPC**: リクエスト・レスポンス型（通常のAPI呼び出し）
- **Server Streaming**: サーバーから継続的にデータを送信（ゲームイベント配信）
- **HTTP/JSON対応**: ブラウザから直接呼び出し可能（プロキシ不要）
- **CORS対応**: クロスオリジンリクエストをサポート

### Protocol Buffers

型安全なAPIスキーマ定義。

- **言語非依存**: Go と TypeScript で同じ型定義を共有
- **バイナリ高速**: JSON よりも高速・軽量（オプション）
- **後方互換性**: フィールドの追加・削除に柔軟に対応
- **自動コード生成**: protoc でクライアント・サーバーコードを生成

### API構成

#### Unary RPC（通常の API）

- `Register`, `Login`: 認証
- `GetAllCards`, `CreateCard`: カード管理
- `CreateDeck`, `GetDecks`, `UpdateDeck`, `DeleteDeck`: デッキ管理
- `CreateGame`, `JoinGame`, `PerformMulligan`: ゲーム開始
- `PlayCard`, `Attack`, `EndTurn`: ゲームアクション
- `GetGameState`: 現在のゲーム状態取得

#### Server Streaming RPC

- `StreamGameEvents`: ゲームイベントのリアルタイム配信
  - カードプレイ、攻撃、ダメージ、ターン終了等のイベント
  - 対戦相手のアクションを即座に反映

## フロントエンド

### 技術スタック

- **React 18 + TypeScript**: コンポーネントベースのUI
- **Vite 5**: 高速な開発サーバー・ビルドツール
- **React Router v7**: ルーティング
- **@connectrpc/connect-web**: Connect-RPCクライアント
- **@bufbuild/protobuf**: Protocol Buffers ランタイム
- **Biome**: Linter + Formatter（ESLint + Prettier の代替）

### 主要コンポーネント

#### ページ

- `Game.tsx`: ゲームメイン画面
- `Admin.tsx`: カード・デッキ管理画面

#### ゲーム関連

- `GameBoard.tsx`: ゲームボード全体（フィールド、手札、プレイヤー情報）
- `GameSetup.tsx`: ゲーム開始画面（デッキ選択、対戦相手入力）
- `MulliganModal.tsx`: マリガン画面
- `HandCard.tsx`: 手札のカード
- `UnitCard.tsx`: フィールドのユニットカード
- `PlayerInfo.tsx`: プレイヤー情報（HP、マナ等）

#### 共通

- `ConnectionStatus.tsx`: 接続状態表示
- `ErrorBoundary.tsx`: エラーハンドリング
- `LoadingSpinner.tsx`: ローディング表示

### カスタムフック

状態管理とロジックをフックに分離。

- `useGameState.ts`: ゲーム状態管理
- `useGameActions.ts`: ゲームアクション（カードプレイ、攻撃等）
- `useMulligan.ts`: マリガン処理
- `useMatchmaking.ts`: マッチメイキング
- `useDeckList.ts`: デッキ一覧管理

### API クライアント

- `lib/api-client.ts`: Connect-RPCクライアント初期化
  - トランスポート設定（HTTP/JSON）
  - ベースURL設定
  - インターセプター（認証ヘッダー追加）
- `lib/auth.ts`: 認証ヘルパー
  - LocalStorageへのトークン保存・取得
  - ログイン・ログアウト処理

### 型定義

- `gen/`: Protocol Buffers から自動生成された型定義
- `types/`: アプリケーション固有の型定義
  - `components.ts`: コンポーネント用の型
  - `deck.ts`: デッキ関連の型

### ビルド・開発

- **開発サーバー**: `pnpm dev`（ホットリロード対応）
- **ビルド**: `pnpm build`（TypeScript + Vite）
- **Linter**: `pnpm lint`（Biome）
- **Formatter**: `pnpm format`（Biome）

## データベース設計

### テーブル構成

#### users

ユーザー情報。

- `id`: UUID（主キー）
- `username`: ユーザー名（ユニーク）
- `password_hash`: パスワードハッシュ（bcrypt）
- `created_at`, `updated_at`

#### cards

カードマスターデータ。

- `id`: カードID（主キー）
- `name`: カード名
- `type`: カードタイプ（Unit / Spell / Leader）
- `cost`: マナコスト
- `attack`: 攻撃力（Unitのみ）
- `health`: 体力（Unitのみ）
- `effect`: 効果（JSON）
- `traits`: 特性（JSON配列）
- `description`: 説明文

#### decks

デッキ情報。

- `id`: デッキID（主キー）
- `user_id`: ユーザーID（外部キー）
- `name`: デッキ名
- `created_at`, `updated_at`

#### deck_cards

デッキとカードの中間テーブル。

- `id`: レコードID（主キー）
- `deck_id`: デッキID（外部キー）
- `card_id`: カードID（外部キー）
- `quantity`: 枚数

### リレーション

```
users (1) ---< (N) decks
decks (1) ---< (N) deck_cards >--- (1) cards
```

## セキュリティ

### 認証・認可

- **JWT**: JSON Web Token による認証
  - HS256 アルゴリズム
  - トークン有効期限: 24時間
  - 秘密鍵: 環境変数 `JWT_SECRET` で設定
- **パスワード**: bcrypt によるハッシュ化（コスト: 10）
- **認証インターセプター**: JWT トークンを検証し、ユーザーIDをコンテキストに設定

### CORS

- **許可オリジン**: フロントエンドの開発サーバー（`http://localhost:5173`）
- **許可ヘッダー**: `Content-Type`, `Authorization`, `Connect-Protocol-Version`
- **許可メソッド**: `GET`, `POST`, `OPTIONS`

### 入力検証

- **Protocol Buffers**: 型レベルでの検証
- **ハンドラー**: ビジネスルールのバリデーション（デッキ枚数、カード重複等）

## パフォーマンス最適化

### バックエンド

- **接続プーリング**: GORM の DB 接続プール
- **インメモリゲーム状態**: ゲーム中の状態はメモリ上で管理
- **イベント駆動**: ゲームイベントを効率的に配信

### フロントエンド

- **コード分割**: React Router によるページごとのコード分割
- **Vite**: 高速な HMR（Hot Module Replacement）
- **型安全**: TypeScript + Protocol Buffers による型安全性

## 今後の拡張予定

- **ランク戦**: レーティングシステム
- **リプレイ機能**: ゲーム履歴の保存・再生
- **フレンド機能**: フレンド対戦
- **チャット機能**: 対戦中のチャット
- **カスタムカード**: ユーザーがカードを作成可能に
- **AI対戦**: CPU対戦モード
- **モバイル対応**: レスポンシブデザイン、PWA化


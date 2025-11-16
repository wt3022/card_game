# カードゲーム - プロジェクトドキュメント

## 目次

1. [概要](#概要)
2. [アーキテクチャ](#アーキテクチャ)
3. [ディレクトリ構造](#ディレクトリ構造)
4. [クリーンアーキテクチャの層](#クリーンアーキテクチャの層)
5. [主要コンポーネント](#主要コンポーネント)
6. [ゲームルール](#ゲームルール)
7. [セットアップ](#セットアップ)
8. [開発ガイド](#開発ガイド)
9. [API仕様](#api仕様)

---

## 概要

このプロジェクトは、**クリーンアーキテクチャ**を採用したターン制のカードゲームです。プレイヤーはユニットを召喚し、スペルを使用し、敵プレイヤーを倒すことを目指します。

### 技術スタック

**バックエンド**
- **言語**: Go 1.25.4
- **通信プロトコル**: Protocol Buffers + Connect-Go
- **アーキテクチャ**: Clean Architecture
- **サーバー**: HTTP/1.1 + HTTP/2 (h2c)

**フロントエンド**
- **フレームワーク**: React 18
- **言語**: TypeScript
- **ビルドツール**: Vite
- **APIクライアント**: Connect-Web
- **パッケージマネージャー**: pnpm

### 主要な特徴

- ✅ ブラウザから直接通信可能（プロキシ不要）
- ✅ HTTP/JSON & Protocol Buffers の両方に対応
- ✅ リアルタイムイベントストリーミング
- ✅ CORS対応
- ✅ クリーンアーキテクチャによる保守性の高い設計

---

## アーキテクチャ

このプロジェクトは**クリーンアーキテクチャ（Clean Architecture）**を採用しています。依存関係は常に外側から内側に向かうように設計されています。

```
┌────────────────────────────────────────────────────────────┐
│                    Infrastructure                           │
│              (event, broadcaster, logger)                   │
└────────────────────────────────────────────────────────────┘
                           ↑
┌────────────────────────────────────────────────────────────┐
│                    Adapter / Interface                      │
│         (connect handler, converter, proto)                 │
└────────────────────────────────────────────────────────────┘
                           ↑
┌────────────────────────────────────────────────────────────┐
│                   Application Service                       │
│                 (GameService - orchestration)               │
└────────────────────────────────────────────────────────────┘
                           ↑
┌────────────────────────────────────────────────────────────┐
│                      Use Cases                              │
│    (Engine, Combat, Effect, Card Play, Turn Management)    │
└────────────────────────────────────────────────────────────┘
                           ↑
┌────────────────────────────────────────────────────────────┐
│                   Core / Domain                             │
│        (Entity, Value Objects, Domain Logic)                │
└────────────────────────────────────────────────────────────┘
```

### 依存性のルール

1. **内側の層は外側の層を知らない**
2. **外側の層は内側の層に依存する**
3. **ビジネスロジックはCoreとUseCaseに集中**
4. **外部とのやり取りはAdapterとInfrastructureで実装**

---

## ディレクトリ構造

```
card_game/
├── api/                          # API定義・自動生成コード
│   ├── proto/                    # Protocol Buffers定義
│   │   ├── common.proto          # 共通メッセージ定義
│   │   └── game.proto            # ゲームサービス定義
│   └── gen/                      # 自動生成されたコード
│       └── proto/cardgame/v1/
│           ├── common.pb.go
│           ├── game.pb.go
│           ├── game_grpc.pb.go
│           └── cardgamev1connect/
│               └── game.connect.go
│
├── cmd/                          # エントリーポイント
│   └── connect-server/
│       └── main.go               # Connect-Goサーバー起動
│
├── internal/                     # 内部実装
│   ├── core/                     # ドメイン層（最内層）
│   │   ├── entity/               # ドメインエンティティ
│   │   │   ├── card.go           # カード
│   │   │   ├── player.go         # プレイヤー
│   │   │   ├── unit.go           # ユニット
│   │   │   ├── effect.go         # 効果定義
│   │   │   ├── enchantment.go    # エンチャント（永続効果）
│   │   │   ├── trait.go          # 特性（Rush、Guardianなど）
│   │   │   ├── combat_result.go  # 戦闘結果
│   │   │   ├── event.go          # ゲームイベント
│   │   │   ├── constants.go      # 定数
│   │   │   ├── errors.go         # ドメインエラー
│   │   │   └── value_objects.go  # 値オブジェクト
│   │   ├── port/                 # インターフェース定義
│   │   │   ├── game_state.go     # ゲーム状態インターフェース
│   │   │   ├── event.go          # イベントインターフェース
│   │   │   └── logger.go         # ロガーインターフェース
│   │   └── usecase/              # ユースケース層
│   │       ├── engine.go         # メインエンジン
│   │       ├── card_play.go      # カードプレイロジック
│   │       ├── combat/           # 戦闘システム
│   │       │   ├── combat.go
│   │       │   ├── destruction.go
│   │       │   └── target.go
│   │       ├── effect/           # 効果処理システム
│   │       │   ├── processor.go
│   │       │   ├── conditions.go
│   │       │   ├── operators.go
│   │       │   ├── targets.go
│   │       │   └── atomic/       # アトミック効果
│   │       │       ├── damage.go
│   │       │       ├── destroy.go
│   │       │       ├── draw.go
│   │       │       ├── heal.go
│   │       │       ├── mana.go
│   │       │       ├── modify.go
│   │       │       └── trait.go
│   │       └── game/             # ゲーム状態管理
│   │           ├── state.go
│   │           ├── phase.go
│   │           ├── turn.go
│   │           ├── victory.go
│   │           ├── log.go
│   │           └── print_battlefield.go
│   │
│   ├── application/              # アプリケーション層
│   │   └── service/
│   │       └── game_service.go   # ゲームサービス（orchestration）
│   │
│   ├── adapter/                  # アダプター層
│   │   ├── connect/
│   │   │   └── handler/
│   │   │       └── game_connect_handler.go  # Connect-Goハンドラー
│   │   └── converter/            # プロトコルバッファ変換
│   │       ├── from_proto.go
│   │       └── to_proto.go
│   │
│   └── infrastructure/           # インフラストラクチャ層
│       └── event/
│           ├── bus.go            # イベントバス
│           ├── broadcaster.go    # イベントブロードキャスト
│           └── handlers.go       # イベントハンドラー
│
├── frontend/                     # フロントエンド（React + TypeScript）
│   ├── src/
│   │   ├── components/           # Reactコンポーネント
│   │   │   ├── GameSetup.tsx     # ゲーム開始画面
│   │   │   ├── GameBoard.tsx     # ゲームボード
│   │   │   ├── PlayerInfo.tsx    # プレイヤー情報表示
│   │   │   ├── UnitCard.tsx      # ユニットカード
│   │   │   └── HandCard.tsx      # 手札カード
│   │   ├── lib/
│   │   │   └── api-client.ts     # Connect-Web APIクライアント
│   │   ├── gen/                  # 自動生成された型定義（.gitignore）
│   │   │   ├── common_pb.ts      # 共通型定義
│   │   │   ├── game_pb.ts        # ゲーム型定義
│   │   │   └── game_connect.ts   # Connect-Webサービス
│   │   ├── App.tsx               # メインアプリケーション
│   │   ├── main.tsx              # エントリーポイント
│   │   └── index.css             # グローバルスタイル
│   ├── buf.gen.yaml              # Buf設定（型定義生成）
│   ├── buf.yaml                  # Buf設定
│   ├── index.html                # HTMLエントリー
│   ├── package.json              # npm依存関係
│   ├── pnpm-lock.yaml            # pnpmロック
│   ├── tsconfig.json             # TypeScript設定
│   ├── vite.config.ts            # Vite設定
│   └── README.md                 # フロントエンド説明書
│
├── bin/                          # ビルド成果物
├── scripts/                      # 各種スクリプト
├── go.mod                        # Go モジュール定義
├── go.sum                        # 依存関係ロック
└── Makefile                      # ビルド・開発タスク
```

---

## クリーンアーキテクチャの層

### 1. Core Layer（ドメイン層）- 最内層

**責務**: ビジネスロジックとドメインルールの定義

#### Entity
- `card.go`: カード（Unit, Spell, Leader）
- `player.go`: プレイヤー（HP、マナ、手札、デッキ、墓地、フィールド）
- `unit.go`: 盤面のユニット（攻撃力、守備力、特性、召喚酔い）
- `effect.go`: カード効果の定義
- `trait.go`: 特性（Rush、Guardian、Lifestealなど）
- `enchantment.go`: 永続効果

#### Port（インターフェース）
- `game_state.go`: ゲーム状態操作のインターフェース
- `event.go`: イベント発行のインターフェース
- `logger.go`: ロギングのインターフェース

#### UseCase
- `engine.go`: ユースケースエンジン（統合窓口）
- `card_play.go`: カードプレイロジック
- `combat/`: 戦闘システム（攻撃、破壊、ターゲット選択）
- `effect/`: 効果処理システム（ダメージ、回復、ドロー、バフなど）
- `game/`: ゲーム状態管理（フェーズ、ターン、勝利判定）

**特徴**:
- 外部ライブラリに依存しない
- テスト可能
- ビジネスルールが集中

---

### 2. Application Layer（アプリケーション層）

**責務**: ユースケースのオーケストレーション

#### GameService
- ゲームセッションの管理
- 複数のユースケースの調整
- トランザクション境界の定義
- イベント発行

**主要メソッド**:
```go
CreateGame(gameID, player1Name, player2Name, deck1, deck2)
GetGameState(gameID)
PlayCard(gameID, playerID, cardID, targetID)
ExecuteAttack(gameID, attackerPlayerID, attackerID, targetID)
EndTurn(gameID)
SubscribeToEvents(gameID)
```

---

### 3. Adapter Layer（アダプター層）

**責務**: 外部からの入力を内部形式に変換

#### Connect Handler
- HTTP/gRPCリクエストをアプリケーションサービスに橋渡し
- Protocol Buffersとドメインエンティティの変換
- エラーハンドリング

#### Converter
- `to_proto.go`: ドメインエンティティ → Protocol Buffers
- `from_proto.go`: Protocol Buffers → ドメインエンティティ

---

### 4. Infrastructure Layer（インフラストラクチャ層）

**責務**: 技術的な実装の詳細

#### Event
- `bus.go`: イベントバス
- `broadcaster.go`: イベントのブロードキャスト
- `handlers.go`: イベントハンドラー

**機能**:
- リアルタイムイベント配信
- ゲームイベントのストリーミング
- 購読管理

---

## 主要コンポーネント

### 1. カード（Card）

```go
type Card struct {
    ID         string      // カードID
    Name       string      // カード名
    Type       CardType    // Unit, Spell, Leader
    Cost       int         // マナコスト
    Attack     *int        // 攻撃力（ユニットのみ）
    Defense    *int        // 守備力（ユニットのみ）
    Effect     string      // 効果テキスト
    CardEffect *CardEffect // 実際の効果定義
    Traits     []Trait     // 特性
}
```

**種類**:
- **Unit**: フィールドに配置される戦闘ユニット
- **Spell**: 即座に効果を発揮する魔法
- **Leader**: プレイヤーを代表するカード（未実装）

---

### 2. ユニット（Unit）

```go
type Unit struct {
    CardID         string
    InstanceID     string  // 盤面での一意ID
    Name           string
    Cost           int
    Attack         int
    Defense        int
    CurrentDefense int     // 現在の守備力
    Traits         []Trait
    OwnerID        string
    SummonSickness bool    // 召喚酔い
    HasAttacked    bool    // 攻撃済みフラグ
}
```

---

### 3. プレイヤー（Player）

```go
type Player struct {
    ID              string
    Name            string
    HP              int           // 現在のHP
    MaxHP           int           // 最大HP
    CurrentTurnMana int           // 現在のマナ
    MaxRecoveryMana int           // ターン開始時の回復マナ
    Hand            []Card        // 手札
    Deck            []Card        // デッキ
    Graveyard       []Card        // 墓地
    Field           []Unit        // フィールド
    TimeRemaining   time.Duration // ターン残り時間
}
```

---

### 4. 特性（Trait）

| 特性 | 説明 |
|------|------|
| **Rush** | 召喚酔いを無視して即座に攻撃可能 |
| **Guardian** | 相手の攻撃を強制的に自分に向ける |
| **Lifesteal** | 与えたダメージ分HPを回復 |
| **Quick** | 即座に行動可能 |

---

### 5. ゲームフェーズ

1. **TurnStart**: ターン開始
2. **Resource**: マナ回復
3. **Draw**: カードドロー
4. **Main**: メインフェーズ（カードプレイ、攻撃）
5. **TurnEnd**: ターン終了

---

## ゲームルール

### 基本ルール

1. **初期設定**
   - 各プレイヤーのHP: 20
   - 初期手札: 4枚
   - 初期マナ: 1

2. **マナシステム**
   - 最大マナ: 10
   - ターン開始時にマナが回復
   - 最大回復マナは徐々に増加（上限10）

3. **ターンの流れ**
   ```
   ターン開始 → マナ回復 → カードドロー → メインフェーズ → ターン終了
   ```

4. **勝利条件**
   - 相手プレイヤーのHPを0にする
   - 相手のデッキが尽きる

### 戦闘ルール

1. **攻撃**
   - ユニットは1ターンに1回攻撃可能
   - 召喚したターンは攻撃不可（召喚酔い）
   - Rushトレイトを持つユニットは召喚ターンに攻撃可能

2. **ターゲット選択**
   - 相手プレイヤーまたは相手ユニットを攻撃
   - Guardianトレイト持ちユニットがいる場合は強制的にそれを攻撃

3. **ダメージ計算**
   - 攻撃側の攻撃力 = 防御側へのダメージ
   - 防御側の攻撃力 = 攻撃側へのダメージ（相互ダメージ）
   - 守備力が0以下になったユニットは破壊

---

## セットアップ

### 前提条件

**バックエンド**
- Go 1.25.4以上
- Protocol Buffers Compiler（protoc）

**フロントエンド**
- Node.js 20.x以上
- pnpm 8.x以上

### インストール

#### 1. リポジトリのクローン
   ```bash
   git clone <repository-url>
   cd card_game
   ```

#### 2. バックエンドのセットアップ

**依存関係のインストール**
   ```bash
   make deps
   ```

**必要なツールのインストール**
   ```bash
   make install-tools
   ```

**Protocol Buffersからコード生成**
   ```bash
   make proto
   ```

**サーバーのビルド**
   ```bash
   make build-connect
   ```

#### 3. フロントエンドのセットアップ

**フロントエンドディレクトリに移動**
```bash
cd frontend
```

**依存関係のインストール**
```bash
pnpm install
```

**型定義の生成（protoファイルから自動生成）**
```bash
pnpm run proto:generate
```

### サーバー起動

#### バックエンド（ターミナル1）

```bash
make run-connect
```

または

```bash
./bin/connect-server
```

サーバーは `http://localhost:8080` で起動します。

#### フロントエンド（ターミナル2）

```bash
cd frontend
pnpm run dev
```

フロントエンドは `http://localhost:3000` で起動します。

### アクセス

ブラウザで `http://localhost:3000` にアクセスしてゲームを開始できます。

---

## 開発ガイド

### Makefileコマンド

| コマンド | 説明 |
|----------|------|
| `make help` | ヘルプを表示 |
| `make install-tools` | 必要なツールをインストール |
| `make proto` | protoファイルからコード生成 |
| `make build-connect` | Connect-Goサーバーをビルド |
| `make run-connect` | サーバーを起動 |
| `make test` | テストを実行 |
| `make clean` | 生成ファイルを削除 |
| `make deps` | 依存関係を更新 |
| `make dev` | 開発用: proto生成→ビルド→実行 |

### 新しい機能の追加

#### 1. 新しいエンティティを追加する場合

1. `internal/core/entity/` に新しいファイルを作成
2. ドメインロジックを実装
3. 必要に応じて `port/` にインターフェースを定義

#### 2. 新しいユースケースを追加する場合

1. `internal/core/usecase/` に新しいファイルを作成
2. `engine.go` にメソッドを追加
3. テストを追加

#### 3. 新しいAPIエンドポイントを追加する場合

1. `api/proto/game.proto` にメッセージとRPCを定義
2. `make proto` でコード生成
3. `internal/adapter/connect/handler/` にハンドラーを実装
4. `internal/adapter/converter/` に変換関数を追加

### テスト方法

#### ヘルスチェック

```bash
curl http://localhost:8080/health
```

#### ゲームを作成

```bash
curl -X POST http://localhost:8080/cardgame.v1.GameService/CreateGame \
  -H 'Content-Type: application/json' \
  -d '{
    "player1_id": "p1",
    "player1_name": "Alice",
    "player2_id": "p2",
    "player2_name": "Bob"
  }'
```

#### ゲーム状態を取得

```bash
curl -X POST http://localhost:8080/cardgame.v1.GameService/GetGameState \
  -H 'Content-Type: application/json' \
  -d '{
    "game_id": "game-p1-p2",
    "player_id": "p1"
  }'
```

#### カードをプレイ

```bash
curl -X POST http://localhost:8080/cardgame.v1.GameService/PlayCard \
  -H 'Content-Type: application/json' \
  -d '{
    "game_id": "game-p1-p2",
    "player_id": "p1",
    "card_id": "p1-unit-1"
  }'
```

#### ターンを終了

```bash
curl -X POST http://localhost:8080/cardgame.v1.GameService/EndTurn \
  -H 'Content-Type: application/json' \
  -d '{
    "game_id": "game-p1-p2",
    "player_id": "p1"
  }'
```

---

## API仕様

### GameService

#### CreateGame
ゲームセッションを新規作成

**Request**:
```protobuf
message CreateGameRequest {
  string player1_id = 1;
  string player1_name = 2;
  string player2_id = 3;
  string player2_name = 4;
}
```

**Response**:
```protobuf
message CreateGameResponse {
  string game_id = 1;
  GameState game_state = 2;
}
```

---

#### GetGameState
ゲームの現在の状態を取得

**Request**:
```protobuf
message GetGameStateRequest {
  string game_id = 1;
  string player_id = 2;
}
```

**Response**:
```protobuf
message GetGameStateResponse {
  GameState game_state = 1;
}
```

---

#### PlayCard
カードをプレイ（ユニット召喚またはスペル使用）

**Request**:
```protobuf
message PlayCardRequest {
  string game_id = 1;
  string player_id = 2;
  string card_id = 3;
  optional string target_id = 4;  // スペルのターゲット
}
```

**Response**:
```protobuf
message PlayCardResponse {
  bool success = 1;
  string message = 2;
  GameState game_state = 3;
}
```

---

#### ExecuteAttack
ユニットで攻撃を実行

**Request**:
```protobuf
message ExecuteAttackRequest {
  string game_id = 1;
  string player_id = 2;
  string attacker_id = 3;
  optional string target_id = 4;  // nullの場合はプレイヤー本体
}
```

**Response**:
```protobuf
message ExecuteAttackResponse {
  bool success = 1;
  string message = 2;
  AttackResult result = 3;
  GameState game_state = 4;
}
```

---

#### EndTurn
現在のプレイヤーのターンを終了

**Request**:
```protobuf
message EndTurnRequest {
  string game_id = 1;
  string player_id = 2;
}
```

**Response**:
```protobuf
message EndTurnResponse {
  bool success = 1;
  string message = 2;
  GameState game_state = 3;
}
```

---

#### StreamGameEvents
ゲームイベントをストリーミング（双方向）

**Request Stream**:
```protobuf
message GameEventRequest {
  string game_id = 1;
  string player_id = 2;
}
```

**Response Stream**:
```protobuf
message GameEventResponse {
  GameEvent event = 1;
  GameState game_state = 2;
}
```

---

## アーキテクチャの利点

### 1. 保守性
- ビジネスロジックが明確に分離
- 変更の影響範囲が限定的
- テストが容易

### 2. 拡張性
- 新しい機能の追加が簡単
- 外部システムの変更に強い
- プラグイン的な設計

### 3. テスタビリティ
- 各層が独立してテスト可能
- モックやスタブの作成が簡単
- ユニットテストの作成が容易

### 4. 技術的負債の削減
- ビジネスロジックが技術詳細に依存しない
- フレームワークやライブラリの変更に強い
- 長期的なメンテナンスコストが低い

---

## 今後の拡張

- [ ] データベース永続化（現在はインメモリ）
- [ ] ユーザー認証・認可
- [ ] マッチメイキングシステム
- [ ] リプレイ機能
- [ ] AIプレイヤー
- [ ] カードバランス調整ツール
- [ ] WebSocketによるリアルタイム通信
- [x] フロントエンドクライアント（React + TypeScript）

---

## ライセンス

このプロジェクトは個人開発プロジェクトです。

---

## 詳細ドキュメント

プロジェクトの詳細なドキュメントは `docs/` ディレクトリにあります。

| ドキュメント | 内容 |
|------------|------|
| [📖 ドキュメント目次](docs/INDEX.md) | すべてのドキュメントの目次とガイド |
| [🏗️ アーキテクチャ](docs/ARCHITECTURE.md) | クリーンアーキテクチャの詳細説明 |
| [👨‍💻 開発者ガイド](docs/DEVELOPMENT.md) | 開発環境構築、コーディング規約、実装ガイド |
| [🎮 ゲームデザイン](docs/GAME_DESIGN.md) | ゲームルール、カードシステム、バランス調整 |

---

## 参考資料

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Connect-Go Documentation](https://connectrpc.com/docs/go/getting-started)
- [Protocol Buffers](https://protobuf.dev/)


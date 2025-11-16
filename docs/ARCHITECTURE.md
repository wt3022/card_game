# アーキテクチャ詳細ドキュメント

## 目次

1. [クリーンアーキテクチャの原則](#クリーンアーキテクチャの原則)
2. [層の詳細](#層の詳細)
3. [データフロー](#データフロー)
4. [依存性の管理](#依存性の管理)
5. [設計パターン](#設計パターン)
6. [エラーハンドリング](#エラーハンドリング)

---

## クリーンアーキテクチャの原則

### 1. 依存性の方向

```
Infrastructure → Adapter → Application → UseCase → Entity
     外層          ↓         ↓           ↓         内層
                依存関係は常に内側に向かう
```

**ルール**:
- 内側の層は外側の層を知らない
- 外側の層は内側の層に依存できる
- ビジネスロジックは内側の層に集中

### 2. 依存性逆転の原則（DIP）

内側の層がインターフェース（Port）を定義し、外側の層が実装を提供します。

```go
// Core Layer - ポートを定義
package port

type GameState interface {
    GetPlayerByID(id string) *entity.Player
    ValidateAction(playerID string) error
}

// Application Layer - 実装を提供
package game

type State struct {
    // 実装...
}

func (s *State) GetPlayerByID(id string) *entity.Player {
    // 実装...
}
```

### 3. 単一責任の原則（SRP）

各コンポーネントは1つの責務のみを持ちます。

- **Entity**: ドメインオブジェクトの表現
- **UseCase**: ビジネスロジックの実行
- **Service**: ユースケースのオーケストレーション
- **Handler**: 外部からのリクエスト処理
- **Converter**: データ形式の変換

---

## 層の詳細

### 1. Entity Layer（エンティティ層）

**場所**: `internal/core/entity/`

**責務**:
- ドメインオブジェクトの定義
- ビジネスルールの実装
- 不変条件の保証

**特徴**:
- 外部ライブラリに依存しない純粋なGo構造体
- メソッドはドメインロジックのみ
- 状態変更は自身の責任範囲内のみ

**例**:

```go
// Player - プレイヤーエンティティ
type Player struct {
    ID              string
    HP              int
    CurrentTurnMana int
    Hand            []Card
    // ...
}

// ドメインロジック: マナを消費
func (p *Player) SpendMana(amount int) error {
    if p.CurrentTurnMana < amount {
        return NewErrInsufficientMana(amount, p.CurrentTurnMana)
    }
    p.CurrentTurnMana -= amount
    return nil
}
```

---

### 2. Port Layer（ポート層）

**場所**: `internal/core/port/`

**責務**:
- インターフェースの定義
- 依存性逆転の実現
- 層間の境界定義

**主要インターフェース**:

#### GameState
ゲーム状態に関する操作を定義

```go
type GameState interface {
    // プレイヤー管理
    GetPlayerByID(id string) *entity.Player
    GetCurrentPlayer() *entity.Player
    GetOpponentPlayer(playerID string) *entity.Player
    
    // バリデーション
    ValidateAction(playerID string) error
    
    // フェーズ管理
    ExecuteTurnStartPhase(player *entity.Player)
    ExecuteDrawPhase(player *entity.Player)
    ExecuteResourceGainPhase(player *entity.Player)
    ExecuteTurnEndPhase(player *entity.Player)
    
    // ゲーム制御
    SwitchTurn()
    IncrementCurrentTurn()
    CheckVictoryConditions()
}
```

#### Logger
ロギング操作を定義

```go
type Logger interface {
    Info(format string, args ...interface{})
    Error(format string, args ...interface{})
    Debug(format string, args ...interface{})
}
```

---

### 3. UseCase Layer（ユースケース層）

**場所**: `internal/core/usecase/`

**責務**:
- ビジネスロジックの実装
- エンティティの操作
- ドメインルールの適用

**構造**:

```
usecase/
├── engine.go           # メインエンジン（統合窓口）
├── card_play.go        # カードプレイロジック
├── combat/             # 戦闘システム
│   ├── combat.go       # 戦闘実行
│   ├── destruction.go  # 破壊処理
│   └── target.go       # ターゲット選択
├── effect/             # 効果処理システム
│   ├── processor.go    # 効果プロセッサー
│   ├── conditions.go   # 条件評価
│   ├── operators.go    # オペレーター
│   ├── targets.go      # ターゲット選択
│   └── atomic/         # アトミック効果
│       ├── damage.go
│       ├── heal.go
│       └── ...
└── game/               # ゲーム状態管理
    ├── state.go        # ゲーム状態
    ├── phase.go        # フェーズ管理
    ├── turn.go         # ターン管理
    └── victory.go      # 勝利判定
```

#### Engine（統合窓口）

```go
type Engine struct {
    State           port.GameState
    EffectProcessor *effect.Processor
}

// ユニット召喚
func (e *Engine) SummonUnit(playerID string, cardID string) (*entity.Card, error) {
    return executeSummonUnit(e.State, playerID, cardID)
}

// 攻撃実行
func (e *Engine) ExecuteAttack(action entity.AttackAction) (*entity.CombatResult, error) {
    return combat.ExecuteAttack(e.State, action)
}
```

---

### 4. Application Layer（アプリケーション層）

**場所**: `internal/application/service/`

**責務**:
- ユースケースのオーケストレーション
- トランザクション境界の定義
- イベント発行
- セッション管理

#### GameService

```go
type GameService struct {
    mu               sync.RWMutex
    games            map[string]*GameSession  // ゲームセッション管理
    eventBroadcaster *event.Broadcaster
    logger           port.Logger
}

type GameSession struct {
    State           *game.State
    UsecaseEngine   *usecase.Engine
    EffectProcessor *effect.Processor
}
```

**主要メソッド**:

```go
// ゲーム作成
func (s *GameService) CreateGame(gameID, player1Name, player2Name string, 
                                  deck1, deck2 []entity.Card) error

// カードプレイ
func (s *GameService) PlayCard(ctx context.Context, gameID, playerID, 
                                cardID string, targetID *string) error

// 攻撃実行
func (s *GameService) ExecuteAttack(ctx context.Context, gameID, 
                                     attackerPlayerID, attackerID string, 
                                     targetID *string) (*entity.CombatResult, error)

// ターン終了
func (s *GameService) EndTurn(ctx context.Context, gameID string) error
```

**責務の分離**:
- GameService: セッション管理、イベント発行
- Engine: ビジネスロジック実行
- State: ゲーム状態の保持

---

### 5. Adapter Layer（アダプター層）

**場所**: `internal/adapter/`

**責務**:
- 外部プロトコルとの変換
- リクエストのバリデーション
- エラーマッピング

#### Connect Handler

```go
type GameConnectHandler struct {
    gameService *service.GameService
}

// PlayCard - Connect-Goのハンドラーメソッド
func (h *GameConnectHandler) PlayCard(
    ctx context.Context,
    req *connect.Request[pbv1.PlayCardRequest],
) (*connect.Response[pbv1.PlayCardResponse], error) {
    // 1. リクエストパラメータを取得
    gameID := req.Msg.GetGameId()
    playerID := req.Msg.GetPlayerId()
    cardID := req.Msg.GetCardId()
    
    // 2. アプリケーションサービスに委譲
    err := h.gameService.PlayCard(ctx, gameID, playerID, cardID, targetID)
    if err != nil {
        return nil, mapDomainErrorToConnectError(err)
    }
    
    // 3. レスポンスを作成
    state, _ := h.gameService.GetGameState(gameID)
    resp := &pbv1.PlayCardResponse{
        Success:   true,
        GameState: converter.GameStateToProto(state, playerID),
    }
    
    return connect.NewResponse(resp), nil
}
```

#### Converter

```go
// ドメイン → Proto
func GameStateToProto(state *game.State, viewerID string) *pbv1.GameState {
    return &pbv1.GameState{
        GameId:          state.GameID,
        CurrentPlayerId: state.CurrentPlayerID,
        CurrentTurn:     int32(state.CurrentTurn),
        // ...
    }
}

// Proto → ドメイン
func CardFromProto(proto *pbv1.Card) entity.Card {
    // ...
}
```

---

### 6. Infrastructure Layer（インフラストラクチャ層）

**場所**: `internal/infrastructure/`

**責務**:
- 技術的な実装の詳細
- 外部システムとの連携
- イベント配信

#### Event System

```go
// Broadcaster - イベントブロードキャスター
type Broadcaster struct {
    mu          sync.RWMutex
    subscribers map[string][]chan *GameEvent
}

func (b *Broadcaster) Broadcast(gameID string, event *GameEvent) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    for _, ch := range b.subscribers[gameID] {
        select {
        case ch <- event:
        default:
            // チャネルがブロックされている場合はスキップ
        }
    }
}
```

---

## データフロー

### リクエスト処理の流れ

```
1. HTTP Request
   ↓
2. Connect Handler (Adapter)
   - リクエストをパース
   - Protocol Buffers → ドメインオブジェクト変換
   ↓
3. GameService (Application)
   - セッション管理
   - トランザクション境界
   - 複数のユースケースを調整
   ↓
4. Engine (UseCase)
   - ビジネスロジック実行
   - エンティティ操作
   ↓
5. Entity (Core)
   - ドメインルール適用
   - 状態変更
   ↓
6. GameService (Application)
   - イベント発行
   ↓
7. Connect Handler (Adapter)
   - ドメインオブジェクト → Protocol Buffers変換
   - レスポンス作成
   ↓
8. HTTP Response
```

### 具体例: カードプレイ

```
Client
  ↓ POST /cardgame.v1.GameService/PlayCard
Handler.PlayCard()
  ↓ gameService.PlayCard(gameID, playerID, cardID, targetID)
GameService.PlayCard()
  ↓ session.UsecaseEngine.SummonUnit(playerID, cardID)
Engine.SummonUnit()
  ↓ executeSummonUnit(state, playerID, cardID)
executeSummonUnit()
  ↓ player.PlayCardFromHand(cardID)
  ↓ player.SummonUnit(card, instanceID)
Player (Entity)
  - マナを消費
  - 手札から削除
  - フィールドに配置
  ↓ return
GameService
  ↓ effectProcessor.ProcessTimingEffects(...)
  ↓ broadcastEvent("card_played")
Handler
  ↓ converter.GameStateToProto(state)
  ↓ return Response
Client
```

---

## 依存性の管理

### 依存性注入（DI）

```go
// main.go
func main() {
    // 1. 最内層から初期化
    logger := port.NewConsoleLogger()
    
    // 2. アプリケーション層
    gameService := service.NewGameService(logger)
    
    // 3. アダプター層
    connectHandler := handler.NewGameConnectHandler(gameService)
    
    // 4. インフラストラクチャ層（HTTPサーバー）
    mux := http.NewServeMux()
    path, handlerFunc := cardgamev1connect.NewGameServiceHandler(connectHandler)
    mux.Handle(path, handlerFunc)
    
    http.ListenAndServe(":8080", mux)
}
```

### インターフェースによる抽象化

```go
// Core Layer - インターフェース定義
type Logger interface {
    Info(format string, args ...interface{})
}

// Infrastructure Layer - 実装
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(format string, args ...interface{}) {
    log.Printf(format, args...)
}

// テスト用 - モック実装
type MockLogger struct {
    messages []string
}

func (l *MockLogger) Info(format string, args ...interface{}) {
    l.messages = append(l.messages, fmt.Sprintf(format, args...))
}
```

---

## 設計パターン

### 1. Repository Pattern（今後の拡張）

現在はインメモリですが、将来的にはリポジトリパターンを導入できます。

```go
// Port Layer
type GameRepository interface {
    Save(game *game.State) error
    FindByID(gameID string) (*game.State, error)
    Delete(gameID string) error
}

// Infrastructure Layer
type PostgresGameRepository struct {
    db *sql.DB
}

func (r *PostgresGameRepository) Save(game *game.State) error {
    // PostgreSQLに保存
}
```

### 2. Strategy Pattern

効果処理システムはStrategyパターンを採用しています。

```go
// 効果定義
type EffectDefinition struct {
    Type         EffectType
    Value        int
    TargetType   TargetType
    Condition    *Condition
}

// 効果プロセッサー
type Processor struct {
    state port.GameState
}

func (p *Processor) ProcessEffectDefinition(def EffectDefinition, ...) error {
    switch def.Type {
    case EffectTypeDamage:
        return atomic.ApplyDamage(p.state, targetID, def.Value)
    case EffectTypeHeal:
        return atomic.ApplyHeal(p.state, targetID, def.Value)
    // ...
    }
}
```

### 3. Observer Pattern

イベントシステムはObserverパターンを採用しています。

```go
// イベント発行
gameService.broadcastEvent(gameID, &event.GameEvent{
    EventType: "card_played",
    Message:   "Player played a card",
})

// イベント購読
eventChan, _ := gameService.SubscribeToEvents(gameID)
for event := range eventChan {
    // イベント処理
}
```

### 4. Facade Pattern

Engineクラスは複数のユースケースへのファサードとして機能します。

```go
type Engine struct {
    State           port.GameState
    EffectProcessor *effect.Processor
}

func (e *Engine) SummonUnit(...) { /* combat パッケージに委譲 */ }
func (e *Engine) ExecuteAttack(...) { /* combat パッケージに委譲 */ }
func (e *Engine) UseSpell(...) { /* card_play に委譲 */ }
```

---

## エラーハンドリング

### ドメインエラー

エンティティ層でドメイン固有のエラーを定義します。

```go
// entity/errors.go
type ErrorCategory string

const (
    ErrorCategoryNotFound       ErrorCategory = "NOT_FOUND"
    ErrorCategoryInvalidInput   ErrorCategory = "INVALID_INPUT"
    ErrorCategoryPrecondition   ErrorCategory = "PRECONDITION"
    ErrorCategoryConflict       ErrorCategory = "CONFLICT"
    ErrorCategoryInternal       ErrorCategory = "INTERNAL"
)

type DomainError struct {
    category ErrorCategory
    message  string
}

func NewErrNotFound(entityType, id string) error {
    return &DomainError{
        category: ErrorCategoryNotFound,
        message:  fmt.Sprintf("%s not found: %s", entityType, id),
    }
}

func NewErrInsufficientMana(required, available int) error {
    return &DomainError{
        category: ErrorCategoryPrecondition,
        message:  fmt.Sprintf("insufficient mana: required %d, available %d", required, available),
    }
}
```

### エラーマッピング

アダプター層でドメインエラーをプロトコル固有のエラーに変換します。

```go
func mapDomainErrorToConnectError(err error) error {
    domainErr, ok := entity.AsDomainError(err)
    if !ok {
        return connect.NewError(connect.CodeInternal, err)
    }
    
    switch domainErr.Category() {
    case entity.ErrorCategoryNotFound:
        return connect.NewError(connect.CodeNotFound, err)
    case entity.ErrorCategoryInvalidInput:
        return connect.NewError(connect.CodeInvalidArgument, err)
    case entity.ErrorCategoryPrecondition:
        return connect.NewError(connect.CodeFailedPrecondition, err)
    case entity.ErrorCategoryConflict:
        return connect.NewError(connect.CodeAlreadyExists, err)
    default:
        return connect.NewError(connect.CodeInternal, err)
    }
}
```

---

## まとめ

### クリーンアーキテクチャの利点

1. **テスタビリティ**
   - 各層を独立してテスト可能
   - モックやスタブの作成が容易

2. **保守性**
   - ビジネスロジックが明確に分離
   - 変更の影響範囲が限定的

3. **拡張性**
   - 新機能の追加が容易
   - 外部システムの変更に強い

4. **技術的独立性**
   - フレームワークやライブラリの変更に強い
   - ビジネスロジックが技術詳細に依存しない

### ベストプラクティス

1. **依存関係は常に内側に向ける**
2. **インターフェースで抽象化する**
3. **ビジネスロジックはCoreとUseCaseに集中**
4. **各層の責務を明確にする**
5. **ドメインエラーを適切に設計する**

---

このアーキテクチャにより、長期的に保守可能で拡張性の高いシステムを実現しています。


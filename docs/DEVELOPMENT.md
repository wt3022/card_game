# 開発者ガイド

## 目次

1. [開発環境のセットアップ](#開発環境のセットアップ)
2. [コーディング規約](#コーディング規約)
3. [新機能の追加方法](#新機能の追加方法)
4. [テストの書き方](#テストの書き方)
5. [デバッグ方法](#デバッグ方法)
6. [トラブルシューティング](#トラブルシューティング)

---

## 開発環境のセットアップ

### 必要なツール

1. **Go**: 1.25.4以上
2. **Protocol Buffers Compiler**: protoc
3. **Make**: ビルドツール
4. **Git**: バージョン管理

### 初期セットアップ

```bash
# 1. リポジトリをクローン
git clone <repository-url>
cd card_game

# 2. 依存関係をインストール
make deps

# 3. 開発ツールをインストール
make install-tools

# 4. Protocol Buffersからコード生成
make proto

# 5. ビルド確認
make build-connect

# 6. テスト実行
make test
```

### エディタの設定

#### VS Code

推奨拡張機能:
- Go（Go Team at Google）
- vscode-proto3（zxh404）
- Better Comments

settings.json:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

---

## コーディング規約

### ディレクトリ構造の原則

```
internal/
├── core/           # ドメイン層（外部依存なし）
│   ├── entity/     # エンティティ
│   ├── port/       # インターフェース
│   └── usecase/    # ユースケース
├── application/    # アプリケーション層
├── adapter/        # アダプター層
└── infrastructure/ # インフラストラクチャ層
```

### 命名規則

#### パッケージ名
- 小文字のみ
- 単数形を使用
- 短く、明確に

```go
// Good
package entity
package usecase
package handler

// Bad
package entities
package UseCases
package game_handler
```

#### 構造体名
- PascalCase
- 名詞を使用
- パッケージ名を繰り返さない

```go
// Good
type Card struct { }
type Player struct { }

// Bad（パッケージがentityの場合）
type EntityCard struct { }
```

#### メソッド名
- PascalCase
- 動詞で始める

```go
// Good
func (p *Player) DrawCard() (*Card, error)
func (p *Player) SpendMana(amount int) error

// Bad
func (p *Player) card_draw() (*Card, error)
```

#### インターフェース名
- PascalCase
- "-er"で終わる（可能な場合）

```go
// Good
type Logger interface { }
type GameState interface { }

// Bad
type ILogger interface { }
type LoggerInterface interface { }
```

### コメント規約

#### パッケージコメント
各パッケージの最初のファイルに記載

```go
// Package entity ドメインエンティティを定義するパッケージ
//
// このパッケージはビジネスロジックの中核を担い、
// 外部ライブラリに依存しません。
package entity
```

#### 公開APIのコメント
すべての公開型、関数、メソッドにコメントを記載

```go
// Player プレイヤーの状態とリソースを管理します
type Player struct {
    ID   string
    Name string
    HP   int
}

// DrawCard デッキから1枚カードを引きます
//
// デッキが空の場合はエラーを返します。
func (p *Player) DrawCard() (*Card, error) {
    // ...
}
```

#### セクションコメント
関連するメソッドをグループ化

```go
// ========================================
// マナ管理
// ========================================

// SpendMana マナを消費します
func (p *Player) SpendMana(amount int) error { }

// RecoverMana マナを回復します
func (p *Player) RecoverMana() { }
```

### エラーハンドリング

#### エラーの作成

```go
// ドメインエラーを使用
return entity.NewErrNotFound("player", playerID)
return entity.NewErrInsufficientMana(required, available)

// カスタムエラーメッセージ
return fmt.Errorf("failed to process effect: %w", err)
```

#### エラーのチェック

```go
// Good
if err != nil {
    return nil, fmt.Errorf("failed to draw card: %w", err)
}

// Bad（エラーを無視）
card, _ := player.DrawCard()
```

### 並行処理

#### Mutexの使用

```go
type GameService struct {
    mu    sync.RWMutex
    games map[string]*GameSession
}

// 読み取り専用
func (s *GameService) GetGameState(gameID string) (*game.State, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // ...
}

// 書き込み
func (s *GameService) PlayCard(...) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ...
}
```

---

## 新機能の追加方法

### 1. 新しいカード効果を追加する

#### ステップ1: エンティティ層で効果タイプを定義

```go
// internal/core/entity/effect.go
const (
    EffectTypeDamage  EffectType = "DAMAGE"
    EffectTypeHeal    EffectType = "HEAL"
    EffectTypeNewType EffectType = "NEW_TYPE"  // 追加
)
```

#### ステップ2: アトミック効果を実装

```go
// internal/core/usecase/effect/atomic/new_effect.go
package atomic

import (
    "card_game/internal/core/entity"
    "card_game/internal/core/port"
)

// ApplyNewEffect 新しい効果を適用
func ApplyNewEffect(state port.GameState, targetID string, value int) error {
    // 実装...
    return nil
}
```

#### ステップ3: プロセッサーに追加

```go
// internal/core/usecase/effect/processor.go
func (p *Processor) ProcessEffectDefinition(def entity.EffectDefinition, ...) error {
    switch def.Type {
    case entity.EffectTypeNewType:
        return atomic.ApplyNewEffect(p.state, targetID, def.Value)
    // ...
    }
}
```

#### ステップ4: テストを追加

```go
// internal/core/usecase/effect/atomic/new_effect_test.go
func TestApplyNewEffect(t *testing.T) {
    // テスト実装
}
```

---

### 2. 新しいAPIエンドポイントを追加する

#### ステップ1: Protocol Buffersに定義を追加

```protobuf
// api/proto/game.proto
service GameService {
  rpc CreateGame(CreateGameRequest) returns (CreateGameResponse);
  rpc NewEndpoint(NewRequest) returns (NewResponse);  // 追加
}

message NewRequest {
  string game_id = 1;
  string player_id = 2;
}

message NewResponse {
  bool success = 1;
  string message = 2;
}
```

#### ステップ2: コード生成

```bash
make proto
```

#### ステップ3: アプリケーションサービスにメソッドを追加

```go
// internal/application/service/game_service.go
func (s *GameService) NewOperation(ctx context.Context, gameID, playerID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    session, exists := s.games[gameID]
    if !exists {
        return entity.NewErrNotFound("game", gameID)
    }
    
    // ビジネスロジック実装
    
    return nil
}
```

#### ステップ4: ハンドラーに実装を追加

```go
// internal/adapter/connect/handler/game_connect_handler.go
func (h *GameConnectHandler) NewEndpoint(
    ctx context.Context,
    req *connect.Request[pbv1.NewRequest],
) (*connect.Response[pbv1.NewResponse], error) {
    gameID := req.Msg.GetGameId()
    playerID := req.Msg.GetPlayerId()
    
    err := h.gameService.NewOperation(ctx, gameID, playerID)
    if err != nil {
        return nil, mapDomainErrorToConnectError(err)
    }
    
    resp := &pbv1.NewResponse{
        Success: true,
        Message: "Operation successful",
    }
    
    return connect.NewResponse(resp), nil
}
```

#### ステップ5: テスト

```bash
# ビルド
make build-connect

# サーバー起動
make run-connect

# 別ターミナルでテスト
curl -X POST http://localhost:8080/cardgame.v1.GameService/NewEndpoint \
  -H 'Content-Type: application/json' \
  -d '{"game_id": "test", "player_id": "p1"}'
```

---

### 3. 新しいエンティティを追加する

#### ステップ1: エンティティ構造体を定義

```go
// internal/core/entity/new_entity.go
package entity

// NewEntity 新しいエンティティ
type NewEntity struct {
    ID   string
    Name string
}

// NewNewEntity コンストラクター
func NewNewEntity(id, name string) *NewEntity {
    return &NewEntity{
        ID:   id,
        Name: name,
    }
}

// Validate バリデーション
func (e *NewEntity) Validate() error {
    if e.ID == "" {
        return NewErrInvalidInput("id", "must not be empty")
    }
    return nil
}
```

#### ステップ2: ユースケースで使用

```go
// internal/core/usecase/new_usecase.go
func ProcessNewEntity(state port.GameState, entity *entity.NewEntity) error {
    if err := entity.Validate(); err != nil {
        return err
    }
    // ロジック実装
    return nil
}
```

---

## テストの書き方

### ユニットテストの基本構造

```go
package entity_test

import (
    "testing"
    "card_game/internal/core/entity"
)

func TestPlayer_SpendMana(t *testing.T) {
    tests := []struct {
        name          string
        initialMana   int
        spendAmount   int
        expectError   bool
        expectedMana  int
    }{
        {
            name:         "正常系: マナ消費成功",
            initialMana:  5,
            spendAmount:  3,
            expectError:  false,
            expectedMana: 2,
        },
        {
            name:         "異常系: マナ不足",
            initialMana:  2,
            spendAmount:  5,
            expectError:  true,
            expectedMana: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            player := &entity.Player{
                ID:              "test-player",
                CurrentTurnMana: tt.initialMana,
            }
            
            err := player.SpendMana(tt.spendAmount)
            
            if tt.expectError && err == nil {
                t.Errorf("expected error but got none")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if player.CurrentTurnMana != tt.expectedMana {
                t.Errorf("expected mana %d, got %d", tt.expectedMana, player.CurrentTurnMana)
            }
        })
    }
}
```

### モックの作成

```go
// internal/core/port/mock_game_state.go
package port

import "card_game/internal/core/entity"

type MockGameState struct {
    Players map[string]*entity.Player
}

func (m *MockGameState) GetPlayerByID(id string) *entity.Player {
    return m.Players[id]
}

func (m *MockGameState) ValidateAction(playerID string) error {
    if m.Players[playerID] == nil {
        return entity.NewErrNotFound("player", playerID)
    }
    return nil
}
```

### テストの実行

```bash
# すべてのテストを実行
make test

# 特定のパッケージをテスト
go test ./internal/core/entity/...

# カバレッジ付き
go test -cover ./...

# 詳細な出力
go test -v ./...
```

---

## デバッグ方法

### ロギング

```go
// ロガーの使用
logger.Info("Processing card: %s", cardID)
logger.Error("Failed to execute attack: %v", err)
logger.Debug("Current state: %+v", state)
```

### ゲーム状態の表示

```go
// internal/core/usecase/game/print_battlefield.go
func (s *State) PrintBattlefield() {
    // 盤面の状態をコンソールに出力
}

// 使用例
state.PrintBattlefield()
```

### デバッグ用のヘルパー

```go
// デバッグ出力を追加
fmt.Printf("DEBUG: player=%+v\n", player)
fmt.Printf("DEBUG: card effect=%+v\n", card.CardEffect)
```

### Connect-Goのデバッグ

```bash
# 詳細なログ付きでサーバー起動
go run cmd/connect-server/main.go

# curlでリクエスト確認
curl -v -X POST http://localhost:8080/cardgame.v1.GameService/GetGameState \
  -H 'Content-Type: application/json' \
  -d '{"game_id": "game-p1-p2", "player_id": "p1"}'
```

---

## トラブルシューティング

### よくある問題

#### 1. Protocol Buffers生成エラー

**問題**: `make proto` が失敗する

**解決策**:
```bash
# protocがインストールされているか確認
protoc --version

# プラグインを再インストール
make install-tools

# 古い生成ファイルを削除
make clean
make proto
```

#### 2. ビルドエラー

**問題**: `undefined: cardgamev1`

**解決策**:
```bash
# Protocol Buffersを再生成
make proto

# モジュールキャッシュをクリア
go clean -modcache
go mod download
```

#### 3. サーバー起動エラー

**問題**: `address already in use`

**解決策**:
```bash
# ポート8080を使用しているプロセスを確認
lsof -i :8080

# プロセスを終了
kill -9 <PID>
```

#### 4. CORS エラー

**問題**: ブラウザからのリクエストがCORSエラーになる

**解決策**:
main.goのcorsMiddleware設定を確認。すでに `Access-Control-Allow-Origin: *` が設定されているはずです。

---

## ベストプラクティス

### 1. 小さなコミット
機能ごとに小さなコミットを作成

```bash
git add internal/core/entity/new_feature.go
git commit -m "feat: Add new entity for feature X"
```

### 2. ブランチ戦略
feature/機能名でブランチを作成

```bash
git checkout -b feature/add-new-card-effect
```

### 3. コードレビュー前のチェックリスト
- [ ] コードが正しくフォーマットされている（`go fmt`）
- [ ] リンターエラーがない
- [ ] テストが追加されている
- [ ] テストがすべて通過する
- [ ] ドキュメントが更新されている

### 4. パフォーマンス考慮
- 不要なアロケーションを避ける
- スライスの容量を事前に確保
- goroutineのリークに注意

```go
// Good
units := make([]Unit, 0, expectedSize)

// Bad（頻繁な再アロケーション）
units := []Unit{}
```

---

## 参考リソース

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Connect-Go Documentation](https://connectrpc.com/docs/go/getting-started)

---

Happy Coding! 🎮


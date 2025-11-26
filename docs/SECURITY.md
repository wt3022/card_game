# フロントエンド改ざん対策ドキュメント

## 概要

カードゲームのフロントエンドは完全にクライアント側で実行されるため、悪意のあるユーザーによる改ざんのリスクがあります。本ドキュメントでは、実装されているセキュリティ対策と追加で必要な対策について説明します。

## 現在実装されているセキュリティ対策

### 1. **サーバー側での完全な状態管理**

すべてのゲーム状態はサーバー側で管理されており、クライアントは状態の「表示」のみを担当します。

```
クライアント: 表示のみ
     ↓ アクションリクエスト
サーバー: 状態管理 + 検証 + 実行
     ↓ 状態更新通知
クライアント: 表示更新
```

**実装箇所:**
- `internal/application/service/game_service.go`: GameSession管理
- `internal/core/usecase/engine.go`: ゲームロジック実行

### 2. **すべてのアクションのサーバー側検証**

クライアントからのすべてのアクションは、サーバー側で厳密に検証されます。

**検証項目:**
- ターン順序の確認
- マナコストの確認
- カードの所有権確認
- ユニットの攻撃可能状態確認
- ターゲットの有効性確認

**実装箇所:**
- `internal/core/usecase/card_play.go`: カードプレイの検証
- `internal/core/usecase/engine.go`: 攻撃・ターン管理の検証

### 3. **JWT認証によるユーザー識別**

各リクエストはJWTトークンで認証され、プレイヤーIDの偽装を防ぎます。

**実装箇所:**
- `internal/adapter/connect/interceptor/auth_interceptor.go`: JWT検証
- `internal/infrastructure/auth/jwt_provider.go`: トークン生成・検証

### 4. **読み取り専用のクライアント状態**

フロントエンドで受け取るゲーム状態は読み取り専用として扱われ、クライアント側での直接変更は無効です。

## 追加で実装すべきセキュリティ対策

### 1. **レート制限 (Rate Limiting)**

**目的:** 短時間での大量リクエストによるDoS攻撃や不正な操作を防ぐ

**実装方法:**
```go
// アクションごとのレート制限
type RateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
}

func (rl *RateLimiter) CheckLimit(playerID string, action string) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    key := fmt.Sprintf("%s:%s", playerID, action)
    now := time.Now()
    
    // 1秒以内のリクエスト数をカウント
    var recentRequests []time.Time
    for _, t := range rl.requests[key] {
        if now.Sub(t) < time.Second {
            recentRequests = append(recentRequests, t)
        }
    }
    
    if len(recentRequests) >= 10 {
        return errors.New("too many requests")
    }
    
    rl.requests[key] = append(recentRequests, now)
    return nil
}
```

### 2. **アクションタイムスタンプ検証**

**目的:** リプレイ攻撃の防止

**実装方法:**
```go
func ValidateActionTimestamp(timestamp int64) error {
    now := time.Now().Unix()
    diff := now - timestamp
    
    // タイムスタンプが5秒以上古い、または未来の場合は拒否
    if diff > 5 || diff < -1 {
        return errors.New("invalid timestamp")
    }
    
    return nil
}
```

### 3. **サーバー側でのランダム性管理**

**目的:** クライアント側でのランダム値操作を防ぐ

**現状:** 既に実装済み
```go
// internal/core/usecase/engine.go
func (e *Engine) shuffleDeck(deck *entity.Deck) {
    rand.Shuffle(len(deck.Cards), func(i, j int) {
        deck.Cards[i], deck.Cards[j] = deck.Cards[j], deck.Cards[i]
    })
}
```

### 4. **チート検出システム**

**目的:** 不正な操作パターンの検出

**実装方法:**
```go
type CheatDetector struct {
    suspiciousActions map[string]int
    mu                sync.Mutex
}

func (cd *CheatDetector) DetectSuspiciousPattern(playerID string, action string) {
    cd.mu.Lock()
    defer cd.mu.Unlock()
    
    // 疑わしいパターンを検出
    // - 不可能なタイミングでのアクション
    // - 通常あり得ない高速操作
    // - 見えないはずの情報へのアクセス
    
    if cd.isSuspicious(playerID, action) {
        cd.suspiciousActions[playerID]++
        
        if cd.suspiciousActions[playerID] >= 5 {
            cd.flagForReview(playerID)
        }
    }
}
```

### 5. **暗号化通信の強制**

**目的:** 通信の盗聴・改ざん防止

**実装方法:**
- HTTPS/TLS通信の強制
- Strict Transport Security (HSTS) ヘッダーの設定

```go
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}
```

### 6. **ゲーム状態のハッシュ検証**

**目的:** クライアント側での状態改ざんの検出

**実装方法:**
```go
func (gs *GameState) GenerateHash() string {
    data := fmt.Sprintf("%s:%d:%d:%s",
        gs.GameID,
        gs.Turn,
        gs.CurrentPlayerIndex,
        gs.Phase,
    )
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func ValidateGameStateHash(state *GameState, hash string) error {
    expectedHash := state.GenerateHash()
    if expectedHash != hash {
        return errors.New("game state tampered")
    }
    return nil
}
```

### 7. **手札・デッキの可視性制御**

**目的:** 相手の手札やデッキ情報の漏洩防止

**実装方法:**
```go
func FilterGameStateForPlayer(state *GameState, viewerID string) *GameState {
    filtered := state.Clone()
    
    for _, player := range filtered.Players {
        if player.UserID != viewerID {
            // 相手の手札を伏せる
            player.Hand = make([]*Card, len(player.Hand))
            // デッキのカード情報も隠す
            player.Deck = &Deck{CardCount: len(player.Deck.Cards)}
        }
    }
    
    return filtered
}
```

### 8. **アクションログの記録**

**目的:** 不正行為の追跡と証拠保全

**実装方法:**
```go
type ActionLog struct {
    Timestamp time.Time
    GameID    string
    PlayerID  string
    Action    string
    Details   map[string]interface{}
    IPAddress string
}

func LogAction(log *ActionLog) {
    // データベースまたはログファイルに記録
    db.Create(log)
}
```

## フロントエンド側での推奨実装

### 1. **状態の不変性**

```typescript
// React での不変状態管理
const [gameState, setGameState] = useState<Readonly<GameState>>(null)

// 状態更新は常にサーバーからの通知のみ
useEffect(() => {
    const unsubscribe = subscribeToGameEvents(gameId, (newState) => {
        setGameState(Object.freeze(newState))
    })
    return unsubscribe
}, [gameId])
```

### 2. **UIの無効化による操作制限**

```typescript
// ターンではない時やマナが足りない時はUIを無効化
const canPlayCard = (card: Card) => {
    return isMyTurn() && 
           currentPlayer.mana >= card.cost &&
           gameState.phase === 'Main'
}

<button 
    disabled={!canPlayCard(card)}
    onClick={() => playCard(card.id)}
>
    Play Card
</button>
```

### 3. **エラーハンドリング**

```typescript
try {
    await gameClient.playCard({ gameId, playerId, cardId })
} catch (error) {
    if (error.code === 'PERMISSION_DENIED') {
        showError('不正な操作が検出されました')
        // 状態を再同期
        await refreshGameState()
    }
}
```

## セキュリティチェックリスト

### サーバー側
- [x] すべてのゲーム状態をサーバーで管理
- [x] すべてのアクションをサーバーで検証
- [x] JWT認証によるプレイヤー識別
- [ ] レート制限の実装
- [ ] アクションタイムスタンプ検証
- [ ] チート検出システム
- [ ] ゲーム状態のハッシュ検証
- [ ] アクションログの記録
- [ ] セキュリティヘッダーの設定

### フロントエンド側
- [x] 状態は読み取り専用として扱う
- [x] サーバーからの通知でのみ状態更新
- [ ] 状態の不変性を強制
- [ ] 適切なエラーハンドリング
- [ ] UIレベルでの操作制限
- [ ] 定期的な状態再同期

## 脆弱性と対策の優先度

### 高優先度
1. **レート制限**: DoS攻撃の防止
2. **ゲーム状態のハッシュ検証**: 改ざん検出
3. **手札・デッキの可視性制御**: 情報漏洩防止

### 中優先度
4. **チート検出システム**: 不正パターンの検出
5. **アクションログ**: 監査証跡
6. **タイムスタンプ検証**: リプレイ攻撃防止

### 低優先度
7. **追加のセキュリティヘッダー**: 深層防御
8. **フロントエンド不変性強制**: 開発体験向上

## まとめ

現在の実装では、基本的なサーバー側検証と状態管理により、多くの改ざんリスクに対応しています。しかし、より堅牢なシステムにするためには、レート制限、ハッシュ検証、チート検出システムの追加実装が推奨されます。

フロントエンドは「信頼できない」という前提で設計されており、すべての重要な処理はサーバー側で行われています。この設計により、クライアント側のコード改ざんがあっても、ゲームの整合性は保たれます。

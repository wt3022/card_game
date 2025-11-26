# クライアント側セキュリティ実装ガイド

## 概要

このドキュメントでは、フロントエンドでの改ざん対策として実装すべきベストプラクティスを説明します。

## 基本原則

### 1. **ゲーム状態は読み取り専用**

クライアント側でゲーム状態を直接変更してはいけません。すべての変更はサーバーからの通知を通じて行われます。

```typescript
// ❌ BAD: クライアント側で状態を直接変更
const playCard = (card: Card) => {
    gameState.currentPlayer.mana -= card.cost
    gameState.currentPlayer.hand = gameState.currentPlayer.hand.filter(c => c.id !== card.id)
    // ...
}

// ✅ GOOD: サーバーにリクエストして、応答で状態を更新
const playCard = async (card: Card) => {
    try {
        await GameService.playCard(gameId, playerId, card.id)
        // サーバーから通知される新しい状態で自動的に更新される
    } catch (error) {
        handleError(error)
    }
}
```

### 2. **不変性の強制**

TypeScriptの型システムを活用して、状態の不変性を強制します。

```typescript
// domain/models/Game.ts
export type GameState = Readonly<{
    gameId: string
    phase: GamePhase
    currentPlayerIndex: number
    players: ReadonlyArray<Readonly<Player>>
    eventLog: ReadonlyArray<Readonly<GameEvent>>
}>

export type Player = Readonly<{
    userId: string
    username: string
    hp: number
    maxHp: number
    mana: number
    maxMana: number
    hand: ReadonlyArray<Readonly<Card>>
    field: ReadonlyArray<Readonly<Unit>>
    deckCount: number
}>
```

### 3. **サーバー通知による状態更新**

```typescript
// services/GameService.ts の streamGameEvents を使用
useEffect(() => {
    if (!gameId || !playerId) return

    const abortController = new AbortController()

    GameService.streamGameEvents(
        gameId,
        playerId,
        (newState) => {
            // サーバーからの状態更新のみ受け入れる
            setGameState(Object.freeze(newState))
        },
        (error) => {
            console.error('Game event stream error:', error)
            setError(error.message)
        },
        abortController.signal
    )

    return () => {
        abortController.abort()
    }
}, [gameId, playerId])
```

## セキュリティ実装のベストプラクティス

### 1. **UIレベルでの制約**

ユーザーが不正な操作を試みることを防ぐために、UIを適切に無効化します。

```typescript
// hooks/useGameValidation.ts
export const useGameValidation = (gameState: GameState, currentPlayerId: string) => {
    const isMyTurn = () => {
        if (!gameState) return false
        const myPlayer = gameState.players.find(p => p.userId === currentPlayerId)
        if (!myPlayer) return false
        return gameState.currentPlayerIndex === gameState.players.indexOf(myPlayer)
    }

    const canPlayCard = (card: Card) => {
        if (!isMyTurn()) return false
        if (gameState.phase !== 'Main') return false
        
        const myPlayer = gameState.players.find(p => p.userId === currentPlayerId)
        if (!myPlayer) return false
        
        return myPlayer.mana >= card.cost
    }

    const canAttack = (unit: Unit) => {
        if (!isMyTurn()) return false
        if (gameState.phase !== 'Main') return false
        if (unit.attacked) return false
        if (unit.summoned && !unit.traits.includes('Rush')) return false
        
        return true
    }

    return {
        isMyTurn,
        canPlayCard,
        canAttack,
    }
}
```

```tsx
// components/HandCard.tsx
const HandCard = ({ card }: { card: Card }) => {
    const { canPlayCard } = useGameValidation(gameState, currentPlayerId)
    const { playCard, isLoading } = useGameActions(gameId, currentPlayerId, onStateUpdate)

    const handleClick = () => {
        if (!canPlayCard(card)) {
            return // UIレベルで無効化されているのでクリックしても何もしない
        }
        playCard(card.id)
    }

    return (
        <div 
            className={`hand-card ${canPlayCard(card) ? 'playable' : 'disabled'}`}
            onClick={handleClick}
        >
            {/* カードの表示 */}
        </div>
    )
}
```

### 2. **エラーハンドリングと状態再同期**

サーバーがリクエストを拒否した場合、適切にエラーを処理し、状態を再同期します。

```typescript
// services/GameService.ts に追加
export class GameService {
    // ... 既存のコード ...

    static async syncGameState(gameId: string, playerId: string): Promise<GameState> {
        const response = await gameClient.getGameState({ gameId, playerId })
        if (!response.gameState) {
            throw new Error('Failed to sync game state')
        }
        return GameMapper.gameStateToDomain(response.gameState)
    }
}

// hooks/useGameActions.ts
const playCard = useCallback(async (cardId: string, targetId?: string) => {
    setIsLoading(true)
    setError(null)

    try {
        await GameService.playCard(gameId, currentPlayerId, cardId, targetId)
        // サーバーからの通知で自動的に状態更新される
        return true
    } catch (error) {
        if (error instanceof ConnectError) {
            switch (error.code) {
                case Code.PermissionDenied:
                    setError('不正な操作が検出されました')
                    // 状態を再同期
                    const freshState = await GameService.syncGameState(gameId, currentPlayerId)
                    onGameStateUpdate(freshState)
                    break
                case Code.FailedPrecondition:
                    setError('その操作は実行できません')
                    break
                case Code.ResourceExhausted:
                    setError('操作が多すぎます。しばらく待ってから再試行してください')
                    break
                default:
                    setError(`エラー: ${error.message}`)
            }
        } else {
            setError('予期しないエラーが発生しました')
        }
        return false
    } finally {
        setIsLoading(false)
    }
}, [gameId, currentPlayerId, onGameStateUpdate])
```

### 3. **リアルタイム検証**

開発環境では、クライアント側でも同じ検証ロジックを実行して早期にエラーを検出できます。

```typescript
// utils/clientValidation.ts
export const validatePlayCard = (
    gameState: GameState,
    playerId: string,
    cardId: string
): { valid: boolean; reason?: string } => {
    // 開発環境でのみ実行
    if (process.env.NODE_ENV !== 'development') {
        return { valid: true }
    }

    const myPlayer = gameState.players.find(p => p.userId === playerId)
    if (!myPlayer) {
        return { valid: false, reason: 'Player not found' }
    }

    const card = myPlayer.hand.find(c => c.id === cardId)
    if (!card) {
        return { valid: false, reason: 'Card not in hand' }
    }

    if (myPlayer.mana < card.cost) {
        return { valid: false, reason: 'Not enough mana' }
    }

    if (gameState.phase !== 'Main') {
        return { valid: false, reason: 'Wrong phase' }
    }

    const isMyTurn = gameState.players[gameState.currentPlayerIndex].userId === playerId
    if (!isMyTurn) {
        return { valid: false, reason: 'Not your turn' }
    }

    return { valid: true }
}

// hooks/useGameActions.ts で使用
const playCard = useCallback(async (cardId: string, targetId?: string) => {
    // 開発環境でのクライアント側検証
    const validation = validatePlayCard(gameState, currentPlayerId, cardId)
    if (!validation.valid) {
        console.warn(`Client validation failed: ${validation.reason}`)
        // 開発環境では警告のみ、本番環境では検証しない
    }

    // サーバーリクエストは常に実行
    return await GameService.playCard(gameId, currentPlayerId, cardId, targetId)
}, [gameState, gameId, currentPlayerId])
```

### 4. **定期的な状態チェック**

長時間のゲームセッションでは、定期的にサーバーと状態を同期します。

```typescript
// hooks/useGameSync.ts
export const useGameSync = (
    gameId: string,
    playerId: string,
    onStateUpdate: (state: GameState) => void
) => {
    useEffect(() => {
        // 30秒ごとに状態を再同期
        const syncInterval = setInterval(async () => {
            try {
                const freshState = await GameService.syncGameState(gameId, playerId)
                onStateUpdate(freshState)
            } catch (error) {
                console.error('Failed to sync game state:', error)
            }
        }, 30000)

        return () => clearInterval(syncInterval)
    }, [gameId, playerId, onStateUpdate])
}
```

### 5. **デバッグモードでの検証表示**

開発中は、クライアント側の状態とサーバー側の状態の差分を可視化します。

```tsx
// components/DebugPanel.tsx (開発環境のみ)
const DebugPanel = ({ gameState }: { gameState: GameState }) => {
    if (process.env.NODE_ENV !== 'development') {
        return null
    }

    const [serverState, setServerState] = useState<GameState | null>(null)

    useEffect(() => {
        const fetchServerState = async () => {
            const state = await GameService.syncGameState(gameState.gameId, currentPlayerId)
            setServerState(state)
        }
        fetchServerState()
    }, [gameState])

    const differences = useMemo(() => {
        if (!serverState) return []
        // 状態の差分を検出
        return detectDifferences(gameState, serverState)
    }, [gameState, serverState])

    return (
        <div className="debug-panel">
            <h3>Debug Info</h3>
            {differences.length > 0 && (
                <div className="warning">
                    State mismatch detected:
                    <ul>
                        {differences.map((diff, i) => (
                            <li key={i}>{diff}</li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    )
}
```

## セキュリティチェックリスト

### 必須項目
- [x] すべての状態更新はサーバーからの通知経由
- [x] ゲーム状態を`Readonly`型で定義
- [x] UIレベルでの操作制限
- [x] 適切なエラーハンドリング
- [x] 不正な操作時の状態再同期

### 推奨項目
- [ ] 開発環境でのクライアント側検証
- [ ] 定期的な状態同期
- [ ] デバッグパネルでの差分表示
- [ ] 操作ログの記録

### 避けるべきパターン
- ❌ クライアント側で直接状態を変更
- ❌ サーバーの応答を待たずにUIを更新
- ❌ ローカルストレージに機密情報を保存
- ❌ クライアント側のバリデーションのみに依存

## まとめ

クライアント側のコードは「信頼できない」という前提で実装する必要があります。すべての重要な処理はサーバー側で行い、クライアントは表示と操作の受付のみを担当します。

この設計により、クライアント側のコードが改ざんされても、ゲームの整合性は保たれます。

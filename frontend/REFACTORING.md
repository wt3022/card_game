# フロントエンド リファクタリング 概要

このドキュメントでは、フロントエンドのコードベースに対して実施したリファクタリングの内容を説明します。

## 🎯 リファクタリングの目的

- **拡張性の向上**: 新機能の追加を容易に
- **保守性の向上**: コードの可読性と管理のしやすさを改善
- **関心の分離**: ビジネスロジックとUIの明確な分離
- **再利用性の向上**: 共通ロジックのカスタムフック化

## 📁 新しいディレクトリ構造

```
frontend/src/
├── components/         # UIコンポーネント
│   ├── Auth/          # 認証関連コンポーネント
│   ├── CardManagement/ # カード管理コンポーネント
│   └── ...            # ゲーム画面コンポーネント
├── hooks/             # カスタムフック（新規）
│   ├── useGameState.ts
│   ├── useGameActions.ts
│   ├── useMatchmaking.ts
│   └── useMulligan.ts
├── lib/               # ライブラリとユーティリティ
│   ├── api-client.ts  # API クライアント統合（改善）
│   └── auth.ts        # 認証ロジック（簡素化）
├── utils/             # ユーティリティ関数（新規）
│   ├── gameHelpers.ts # ゲームロジックヘルパー
│   └── common.ts      # 共通ユーティリティ
├── constants/         # 定数定義（新規）
│   └── game.ts        # ゲーム関連定数
├── types/             # 型定義（新規）
│   └── components.ts  # コンポーネントの型定義
└── pages/             # ページコンポーネント
    ├── Game.tsx
    └── Admin.tsx
```

## ✨ 主な改善点

### 1. APIクライアントの統合と管理改善

**変更前:**
- 各コンポーネントでクライアントを個別に作成
- 認証ロジックが`auth.ts`に混在

**変更後:**
```typescript
// lib/api-client.ts
export const gameClient = createClient(GameService)
export const authClient = createClient(AuthService)
export const cardManagementClient = createAuthenticatedClient(CardManagementService)
```

**メリット:**
- 一元管理により設定変更が容易
- 認証インターセプターの統一
- テストの際のモック化が容易

### 2. カスタムフックによるロジックの分離

#### `useGameState` - ゲーム状態管理
```typescript
const { gameState, error, updateGameState } = useGameState(gameId, playerId)
```
- リアルタイム更新の購読
- エラー処理の統合
- 自動再接続機能

#### `useGameActions` - ゲームアクション実行
```typescript
const { playCard, executeAttack, endTurn, message } = useGameActions(
  gameId,
  currentPlayerId,
  onGameStateUpdate
)
```
- カードプレイ、攻撃、ターン終了のロジック統合
- ローディング状態の管理
- エラーハンドリングの統一

#### `useMatchmaking` - マッチング処理
```typescript
const { isMatchmaking, matchmakingStatus, joinMatchmaking, cancelMatchmaking } = 
  useMatchmaking()
```
- マッチング状態の管理
- ストリーミングイベントの処理
- エラーハンドリング

#### `useMulligan` - マリガン処理
```typescript
const { performMulligan, isWaitingForOpponent } = useMulligan()
```
- マリガン実行ロジック
- 相手待機状態の管理
- イベント購読の自動化

### 3. ユーティリティ関数の整理

#### `utils/gameHelpers.ts`
```typescript
// プレイヤー情報取得
getCurrentPlayer(gameState, currentPlayerId)
getOpponent(gameState, currentPlayerId)

// 状態判定
isCurrentPlayerTurn(gameState, currentPlayerId)
isMyUnit(player, unitId)

// マリガンヘルパー
getMulliganHand(gameState, playerId)
```

#### `utils/common.ts`
```typescript
// エラー処理
formatErrorMessage(error, defaultMessage)

// 非同期処理
delay(ms)
debounce(func, wait)
```

### 4. 定数の一元化

#### `constants/game.ts`
```typescript
export const TRAIT_LABELS = {
  [Trait.RUSH]: '疾走',
  [Trait.CHARGE]: '突進',
  // ...
}

export const MESSAGE_DISPLAY_DURATION = 3000
export const RECONNECT_DELAY = 3000
```

**メリット:**
- マジックナンバーの排除
- 一括変更が容易
- 可読性の向上

### 5. 型定義の一元化

#### `types/components.ts`
```typescript
export interface GameBoardProps {
  gameState: GameState
  currentPlayerId: string
  onGameStateUpdate: (gameState: GameState) => void
}
```

**メリット:**
- 型の一貫性確保
- インターフェースの明確化
- リファクタリングの容易性

### 6. コンポーネントの簡素化

#### GameBoard（変更前: 366行 → 変更後: 約250行）
- カスタムフックによるロジック分離
- 状態管理の簡素化
- 副作用の整理

#### GameSetup（変更前: 297行 → 変更後: 約150行）
- マッチングロジックの分離
- マリガンロジックの分離
- UIとロジックの明確な分離

## 🔧 使用例

### コンポーネントでのカスタムフック利用

```typescript
function GameBoard({ gameState, currentPlayerId, onGameStateUpdate }: GameBoardProps) {
  // ゲーム状態の購読
  const { gameState: liveGameState } = useGameState(
    gameState.gameId,
    currentPlayerId
  )

  // ゲームアクション
  const { playCard, executeAttack, endTurn, message } = useGameActions(
    gameState.gameId,
    currentPlayerId,
    onGameStateUpdate
  )

  // ヘルパー関数の利用
  const currentPlayer = getCurrentPlayer(gameState, currentPlayerId)
  const opponent = getOpponent(gameState, currentPlayerId)

  // ...
}
```

## 🚀 今後の拡張性

このリファクタリングにより、以下の拡張が容易になりました:

1. **新しいゲーム機能の追加**
   - カスタムフックに新しいアクションを追加
   - ヘルパー関数で複雑なロジックを抽出

2. **状態管理ライブラリの導入**
   - フックの内部実装を変更するだけで対応可能
   - コンポーネントへの影響を最小限に

3. **テストの追加**
   - フックとヘルパー関数は独立してテスト可能
   - モックの作成が容易

4. **パフォーマンス最適化**
   - メモ化の追加が容易
   - 再レンダリングの制御が明確

## 📝 マイグレーションガイド

### 旧コードの置き換え

旧バージョンのファイルは `.old.tsx` として保存されています:
- `GameBoard.old.tsx`
- `GameSetup.old.tsx`

問題が発生した場合は、これらのファイルを参照できます。

### APIクライアントの使用方法変更

**変更前:**
```typescript
const cardClient = useMemo(
  () => createAuthenticatedClient(CardManagementService),
  []
)
```

**変更後:**
```typescript
import { cardManagementClient } from '../lib/api-client'
// 直接使用
await cardManagementClient.listCards({})
```

## 🎉 まとめ

このリファクタリングにより、以下の改善が実現されました:

- ✅ コードの重複を削減
- ✅ 関心の分離を明確化
- ✅ テスタビリティの向上
- ✅ 新機能追加の容易性向上
- ✅ バグの発見と修正が容易に
- ✅ チーム開発での可読性向上

これらの改善により、プロジェクトの長期的な保守性と拡張性が大幅に向上しました。

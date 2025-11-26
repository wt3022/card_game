# カード効果タイミングの表示について

## 概要

カード効果の発動タイミング(`EffectTiming`)をフロントエンドで表示できるようになりました。

## 実装内容

### 1. Protocol Buffers定義の追加

`api/proto/common.proto`に`EffectTiming` enumを追加しました:

```protobuf
enum EffectTiming {
  EFFECT_TIMING_UNSPECIFIED = 0;
  EFFECT_TIMING_IMMEDIATE = 1;      // 即座に発動
  EFFECT_TIMING_ON_SUMMON = 2;      // 召喚時
  EFFECT_TIMING_ON_DESTROY = 3;     // 破壊時
  EFFECT_TIMING_ON_ATTACK = 4;      // 攻撃時
  EFFECT_TIMING_ON_DAMAGED = 5;     // ダメージを受けた時
  EFFECT_TIMING_TURN_START = 6;     // ターン開始時
  EFFECT_TIMING_TURN_END = 7;       // ターン終了時
}
```

`AtomicEffect`メッセージに`timing`フィールドを追加:

```protobuf
message AtomicEffect {
  uint32 id = 1;
  AtomicEffectType type = 2;
  optional TargetSelector target = 3;
  optional int32 value = 4;
  optional string card_id = 5;
  optional Trait trait = 6;
  EffectTiming timing = 7;  // 発動タイミング
}
```

### 2. バックエンドコンバーター

- `internal/adapter/converter/to_proto.go`に`effectTimingToProto()`関数を追加
- `internal/adapter/converter/from_proto.go`に`effectTimingFromProto()`関数を追加
- `atomicEffectToProto()`と`atomicEffectFromProto()`でタイミング情報を変換

### 3. フロントエンドコンポーネント

#### 定数とヘルパー関数

`frontend/src/constants/effectTiming.ts`:
- `EFFECT_TIMING_LABELS`: 日本語ラベル（例: "召喚時"）
- `EFFECT_TIMING_SHORT_LABELS`: 短縮ラベル（例: "召"）
- `EFFECT_TIMING_DESCRIPTIONS`: 説明文
- `EFFECT_TIMING_ICONS`: アイコン（絵文字）
- ヘルパー関数: `getEffectTimingLabel()`, `getEffectTimingIcon()`など

#### UIコンポーネント

`frontend/src/components/EffectTimingBadge.tsx`:
- `<EffectTimingBadge>`: 単一のタイミングバッジを表示
- `<EffectTimingList>`: 複数のタイミングバッジをリスト表示
- タイミングごとに色分け（即座=金色、召喚時=水色、破壊時=赤色など）
- ホバーで詳細説明を表示するツールチップ

## 使用方法

### 基本的な使用例

```tsx
import { EffectTimingBadge, EffectTimingList } from './components/EffectTimingBadge'
import { EffectTiming } from './gen/common_pb'

// 単一のタイミングバッジ
<EffectTimingBadge 
  timing={EffectTiming.ON_SUMMON} 
  showLabel={true}
  showTooltip={true}
/>

// 複数のタイミングバッジ
<EffectTimingList 
  timings={[EffectTiming.ON_SUMMON, EffectTiming.ON_DESTROY]} 
  showLabel={true}
  showTooltip={true}
/>
```

### カード/ユニット表示への統合

カードやユニットに効果タイミングを表示するには、`CardEffect`フィールドから`AtomicEffect`の配列を取得し、各効果の`timing`フィールドを使用します:

```tsx
import { EffectTimingList } from './components/EffectTimingBadge'

function CardDisplay({ card }) {
  // CardEffectからタイミング情報を抽出
  const timings = card.cardEffect?.definitions
    ?.flatMap(def => extractTimingsFromNode(def.root))
    .filter((t, i, arr) => arr.indexOf(t) === i) // 重複除去
    ?? []

  return (
    <div className="card">
      <div className="card-name">{card.name}</div>
      <div className="card-effect">{card.effect}</div>
      {timings.length > 0 && (
        <EffectTimingList timings={timings} />
      )}
    </div>
  )
}

// EffectChainNodeからタイミング情報を再帰的に抽出
function extractTimingsFromNode(node) {
  if (!node) return []
  
  const timings = []
  
  if (node.atomicEffect) {
    timings.push(node.atomicEffect.timing)
  }
  
  if (node.next) {
    timings.push(...extractTimingsFromNode(node.next))
  }
  
  if (node.children) {
    node.children.forEach(child => {
      timings.push(...extractTimingsFromNode(child))
    })
  }
  
  // その他のノードタイプ（thenNode, elseNode, repeatEffect, foreachEffectなど）も同様に処理
  
  return timings
}
```

## 現在の制限事項

### CardEffectデータの取得

現在、ゲームプレイ中の`GameState`に含まれる`Card`および`Unit`メッセージには、`card_effect`フィールドが設定されていません。これは以下の理由によります:

1. **設計の分離**: 
   - `entity.Card`の`Effect`フィールド（文字列）: ゲームプレイ用の表示テキスト
   - `entity.Card`の`CardEffect`フィールド（構造体）: カード管理UI用の詳細定義

2. **データ量の削減**: 
   - ゲームプレイ中は効果の実行結果のみが重要で、詳細な効果定義は不要
   - CardEffect構造体は複雑で大きいため、毎回の状態更新に含めるとオーバーヘッドが大きい

### 対応方法

効果タイミングをゲームプレイ中に表示するには、以下のいずれかの方法が必要です:

#### オプション1: 効果テキストからの推測

効果テキストに含まれるキーワードからタイミングを推測:

```tsx
function guessEffectTiming(effectText: string): EffectTiming[] {
  const timings: EffectTiming[] = []
  
  if (effectText.includes('召喚時')) {
    timings.push(EffectTiming.ON_SUMMON)
  }
  if (effectText.includes('破壊時')) {
    timings.push(EffectTiming.ON_DESTROY)
  }
  // ... 他のパターン
  
  return timings
}
```

**利点**: 実装が簡単、既存のデータで動作  
**欠点**: 不正確、効果テキストの書き方に依存

#### オプション2: カードマスターデータの事前読み込み

ゲーム開始時に全カードのCardEffect情報を取得してキャッシュ:

```tsx
// ゲーム開始時
const cardEffects = await cardManagementClient.getCardEffects()
const cardEffectMap = new Map(
  cardEffects.map(ce => [ce.cardId, ce])
)

// ゲームプレイ中
function getCardEffectTiming(cardId: string): EffectTiming[] {
  const cardEffect = cardEffectMap.get(cardId)
  return extractTimingsFromCardEffect(cardEffect)
}
```

**利点**: 正確な情報、一度の取得で全カードに対応  
**欠点**: 初期ロードが重い、メモリ使用量が増える

#### オプション3: GameStateにタイミング情報を含める（推奨）

`Card`メッセージに簡略化されたタイミング情報を追加:

```protobuf
message Card {
  string id = 1;
  string name = 2;
  CardType type = 3;
  int32 cost = 4;
  optional int32 attack = 5;
  optional int32 defense = 6;
  string effect = 7;
  repeated Trait traits = 8;
  optional CardEffect card_effect = 9;
  repeated EffectTiming effect_timings = 10;  // 追加: 効果タイミングのリスト
}
```

バックエンドで`CardToProto()`変換時にタイミング情報を抽出して設定:

```go
func CardToProto(card *entity.Card) *cardgamev1.Card {
    // ... 既存のコード
    
    // CardEffectからタイミング情報を抽出
    var effectTimings []cardgamev1.EffectTiming
    if card.CardEffect != nil {
        effectTimings = extractEffectTimings(card.CardEffect)
    }
    
    return &cardgamev1.Card{
        // ... 既存のフィールド
        EffectTimings: effectTimings,
    }
}
```

**利点**: 正確、効率的、実装が明確  
**欠点**: protoファイルの変更が必要、バックエンド実装が必要

## 推奨実装

短期的には**オプション1**（テキスト推測）で簡易的に実装し、長期的には**オプション3**（protoへの追加）に移行することを推奨します。

### 実装例（オプション1）

```tsx
// frontend/src/utils/effectTimingUtils.ts
import { EffectTiming } from '../gen/common_pb'

export function extractEffectTimingsFromText(effectText: string): EffectTiming[] {
  if (!effectText) return []
  
  const timings: EffectTiming[] = []
  
  // キーワードマッチング
  const patterns = [
    { pattern: /召喚時|場に出たとき/, timing: EffectTiming.ON_SUMMON },
    { pattern: /破壊時|破壊されたとき/, timing: EffectTiming.ON_DESTROY },
    { pattern: /攻撃時|攻撃したとき/, timing: EffectTiming.ON_ATTACK },
    { pattern: /ダメージを受けたとき/, timing: EffectTiming.ON_DAMAGED },
    { pattern: /ターン開始時/, timing: EffectTiming.TURN_START },
    { pattern: /ターン終了時/, timing: EffectTiming.TURN_END },
  ]
  
  for (const { pattern, timing } of patterns) {
    if (pattern.test(effectText)) {
      timings.push(timing)
    }
  }
  
  // パターンにマッチしない場合は即座発動と推測
  if (timings.length === 0) {
    timings.push(EffectTiming.IMMEDIATE)
  }
  
  return timings
}
```

使用例:

```tsx
import { extractEffectTimingsFromText } from '../utils/effectTimingUtils'
import { EffectTimingList } from './EffectTimingBadge'

function UnitCard({ unit }) {
  const timings = extractEffectTimingsFromText(unit.effect)
  
  return (
    <div className="unit-card">
      {/* ... */}
      {unit.effect && (
        <div className="unit-effect">
          {unit.effect}
          <EffectTimingList timings={timings} showLabel={false} />
        </div>
      )}
    </div>
  )
}
```

## まとめ

- Protocol Buffers定義に`EffectTiming` enumを追加済み
- バックエンドのコンバーターでタイミング情報を変換可能
- フロントエンドに表示用コンポーネント`EffectTimingBadge`を実装済み
- ゲームプレイ中のタイミング表示には追加実装が必要（オプション1〜3）
- 推奨: 短期的にはテキスト推測、長期的にはprotoへのフィールド追加


/**
 * 効果タイミング関連のユーティリティ関数
 */
import { EffectTiming } from '../gen/common_pb'

/**
 * 効果テキストからタイミング情報を推測
 *
 * 注意: これは暫定的な実装です。正確な情報を取得するには、
 * CardEffectフィールドを使用する必要があります。
 */
export function extractEffectTimingsFromText(
  effectText: string,
): EffectTiming[] {
  if (!effectText || effectText.trim() === '') {
    return []
  }

  const timings: EffectTiming[] = []

  // キーワードマッチングパターン
  const patterns: Array<{ pattern: RegExp; timing: EffectTiming }> = [
    {
      pattern: /召喚時|場に出たとき|場に出た時/,
      timing: EffectTiming.ON_SUMMON,
    },
    {
      pattern: /破壊時|破壊されたとき|破壊された時/,
      timing: EffectTiming.ON_DESTROY,
    },
    {
      pattern: /攻撃時|攻撃したとき|攻撃した時/,
      timing: EffectTiming.ON_ATTACK,
    },
    {
      pattern: /ダメージを受けたとき|ダメージを受けた時|被ダメージ時/,
      timing: EffectTiming.ON_DAMAGED,
    },
    {
      pattern: /ターン開始時|自分のターン開始時/,
      timing: EffectTiming.TURN_START,
    },
    {
      pattern: /ターン終了時|自分のターン終了時/,
      timing: EffectTiming.TURN_END,
    },
  ]

  // パターンマッチング
  for (const { pattern, timing } of patterns) {
    if (pattern.test(effectText)) {
      timings.push(timing)
    }
  }

  // マッチするパターンがない場合は即座発動と推測
  // ただし、空の効果テキストの場合は除外
  if (timings.length === 0 && effectText.trim() !== '') {
    timings.push(EffectTiming.IMMEDIATE)
  }

  // 重複を除去
  return Array.from(new Set(timings))
}

/**
 * EffectChainNodeから再帰的にタイミング情報を抽出
 *
 * CardEffectが利用可能な場合に使用します。
 */
export function extractTimingsFromEffectDefinition(
  definition: { root?: unknown } | null | undefined,
): EffectTiming[] {
  if (!definition || !definition.root) {
    return []
  }

  return extractTimingsFromNode(definition.root)
}

/**
 * EffectChainNodeから再帰的にタイミング情報を抽出（内部関数）
 */
function extractTimingsFromNode(node: unknown): EffectTiming[] {
  if (!node || typeof node !== 'object') {
    return []
  }

  const timings: EffectTiming[] = []
  const nodeObj = node as Record<string, unknown>

  // AtomicEffectのタイミングを取得
  const atomicEffect = nodeObj.atomicEffect as
    | { timing?: EffectTiming }
    | undefined
  if (atomicEffect?.timing) {
    timings.push(atomicEffect.timing)
  }

  // 順次実行ノードの次のノードを処理
  if (nodeObj.next) {
    timings.push(...extractTimingsFromNode(nodeObj.next))
  }

  // 並列実行ノードの子ノードを処理
  if (nodeObj.children && Array.isArray(nodeObj.children)) {
    for (const child of nodeObj.children) {
      timings.push(...extractTimingsFromNode(child))
    }
  }

  // 条件分岐ノードの分岐先を処理
  if (nodeObj.thenNode) {
    timings.push(...extractTimingsFromNode(nodeObj.thenNode))
  }
  if (nodeObj.elseNode) {
    timings.push(...extractTimingsFromNode(nodeObj.elseNode))
  }

  // 繰り返しノードの効果を処理
  if (nodeObj.repeatEffect) {
    timings.push(...extractTimingsFromNode(nodeObj.repeatEffect))
  }

  // 反復ノードの効果を処理
  if (nodeObj.foreachEffect) {
    timings.push(...extractTimingsFromNode(nodeObj.foreachEffect))
  }

  // 重複を除去
  return Array.from(new Set(timings))
}

/**
 * CardEffectから全てのタイミング情報を抽出
 */
export function extractTimingsFromCardEffect(
  cardEffect: { definitions?: unknown[] } | null | undefined,
): EffectTiming[] {
  if (!cardEffect || !cardEffect.definitions) {
    return []
  }

  const allTimings: EffectTiming[] = []

  for (const definition of cardEffect.definitions) {
    const timings = extractTimingsFromEffectDefinition(
      definition as { root?: unknown } | null | undefined,
    )
    allTimings.push(...timings)
  }

  // 重複を除去してソート
  return Array.from(new Set(allTimings)).sort((a, b) => a - b)
}

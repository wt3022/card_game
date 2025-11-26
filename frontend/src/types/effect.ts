/**
 * 効果関連の型定義
 */
import type { AtomicEffect, EffectTiming } from '../gen/common_pb'

/**
 * 効果の表示情報
 */
export interface EffectDisplayInfo {
  timing: EffectTiming
  description: string
  timingLabel: string
  timingIcon: string
}

/**
 * 効果を持つカード・ユニットの共通インターフェース
 */
export interface HasEffect {
  effect?: string
  effects?: AtomicEffect[]
}

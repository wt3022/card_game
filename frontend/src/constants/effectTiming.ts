/**
 * 効果発動タイミングの定数と変換関数
 */
import { EffectTiming } from '../gen/common_pb'

/**
 * 効果タイミングの日本語ラベル
 */
export const EFFECT_TIMING_LABELS: Record<EffectTiming, string> = {
  [EffectTiming.UNSPECIFIED]: '不明',
  [EffectTiming.IMMEDIATE]: '即座',
  [EffectTiming.ON_SUMMON]: '召喚時',
  [EffectTiming.ON_DESTROY]: '破壊時',
  [EffectTiming.ON_ATTACK]: '攻撃時',
  [EffectTiming.ON_DAMAGED]: 'ダメージ時',
  [EffectTiming.TURN_START]: 'ターン開始時',
  [EffectTiming.TURN_END]: 'ターン終了時',
}

/**
 * 効果タイミングの短縮ラベル（UI表示用）
 */
export const EFFECT_TIMING_SHORT_LABELS: Record<EffectTiming, string> = {
  [EffectTiming.UNSPECIFIED]: '?',
  [EffectTiming.IMMEDIATE]: '即',
  [EffectTiming.ON_SUMMON]: '召',
  [EffectTiming.ON_DESTROY]: '破',
  [EffectTiming.ON_ATTACK]: '攻',
  [EffectTiming.ON_DAMAGED]: '被',
  [EffectTiming.TURN_START]: '始',
  [EffectTiming.TURN_END]: '終',
}

/**
 * 効果タイミングの説明文
 */
export const EFFECT_TIMING_DESCRIPTIONS: Record<EffectTiming, string> = {
  [EffectTiming.UNSPECIFIED]: '不明なタイミング',
  [EffectTiming.IMMEDIATE]: 'カードを使用した時に即座に発動',
  [EffectTiming.ON_SUMMON]: 'ユニットが場に出た時に発動',
  [EffectTiming.ON_DESTROY]: 'ユニットが破壊された時に発動',
  [EffectTiming.ON_ATTACK]: 'ユニットが攻撃した時に発動',
  [EffectTiming.ON_DAMAGED]: 'ユニットがダメージを受けた時に発動',
  [EffectTiming.TURN_START]: 'ターン開始時に発動',
  [EffectTiming.TURN_END]: 'ターン終了時に発動',
}

/**
 * 効果タイミングのアイコン（絵文字）
 */
export const EFFECT_TIMING_ICONS: Record<EffectTiming, string> = {
  [EffectTiming.UNSPECIFIED]: '❓',
  [EffectTiming.IMMEDIATE]: '⚡',
  [EffectTiming.ON_SUMMON]: '🌟',
  [EffectTiming.ON_DESTROY]: '💥',
  [EffectTiming.ON_ATTACK]: '⚔️',
  [EffectTiming.ON_DAMAGED]: '🩹',
  [EffectTiming.TURN_START]: '🌅',
  [EffectTiming.TURN_END]: '🌙',
}

/**
 * 効果タイミングのラベルを取得
 */
export function getEffectTimingLabel(timing: EffectTiming): string {
  return (
    EFFECT_TIMING_LABELS[timing] ||
    EFFECT_TIMING_LABELS[EffectTiming.UNSPECIFIED]
  )
}

/**
 * 効果タイミングの短縮ラベルを取得
 */
export function getEffectTimingShortLabel(timing: EffectTiming): string {
  return (
    EFFECT_TIMING_SHORT_LABELS[timing] ||
    EFFECT_TIMING_SHORT_LABELS[EffectTiming.UNSPECIFIED]
  )
}

/**
 * 効果タイミングの説明文を取得
 */
export function getEffectTimingDescription(timing: EffectTiming): string {
  return (
    EFFECT_TIMING_DESCRIPTIONS[timing] ||
    EFFECT_TIMING_DESCRIPTIONS[EffectTiming.UNSPECIFIED]
  )
}

/**
 * 効果タイミングのアイコンを取得
 */
export function getEffectTimingIcon(timing: EffectTiming): string {
  return (
    EFFECT_TIMING_ICONS[timing] || EFFECT_TIMING_ICONS[EffectTiming.UNSPECIFIED]
  )
}

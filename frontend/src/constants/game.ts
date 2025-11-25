import { Trait } from '../gen/common_pb'

/**
 * 特性のラベルマッピング
 */
export const TRAIT_LABELS: Record<number, string> = {
  [Trait.RUSH]: '疾走',
  [Trait.CHARGE]: '突進',
  [Trait.WINDFURY]: '疾風',
  [Trait.PIERCE]: '貫通',
  [Trait.GUARDIAN]: '守護',
  [Trait.EFFECT_SHIELD]: '効果盾',
  [Trait.UNTARGETABLE]: '対象不可',
}

/**
 * メッセージ表示時間（ミリ秒）
 */
export const MESSAGE_DISPLAY_DURATION = 3000

/**
 * エラー時の再接続遅延（ミリ秒）
 */
export const RECONNECT_DELAY = 3000

/**
 * マリガン後の待機時間（ミリ秒）
 */
export const MULLIGAN_WAIT_TIME = 300

/**
 * ドメイン層: カードモデル
 * Protoから独立した純粋なドメインモデル
 */

export type CardType = 'Unit' | 'Spell' | 'Leader'

export type Trait =
  | 'Rush'
  | 'Charge'
  | 'Windfury'
  | 'Pierce'
  | 'Guardian'
  | 'EffectShield'
  | 'Untargetable'

export interface Card {
  id: string
  name: string
  type: CardType
  cost: number
  attack?: number
  defense?: number
  effect: string
  traits: Trait[]
  cardEffect?: CardEffect
}

// 効果システム
export type AtomicEffectType =
  | 'DEAL_DAMAGE'
  | 'DEAL_SPLASH'
  | 'RESTORE_HP'
  | 'RESTORE_MANA'
  | 'FULL_RESTORE'
  | 'DRAW_CARD'
  | 'DISCARD_CARD'
  | 'SEARCH_CARD'
  | 'SHUFFLE_DECK'
  | 'MODIFY_ATTACK'
  | 'MODIFY_DEFENSE'
  | 'MODIFY_COST'
  | 'MODIFY_MAX_HP'
  | 'SUMMON_UNIT'
  | 'DESTROY_UNIT'
  | 'RETURN_TO_HAND'
  | 'RETURN_TO_DECK'
  | 'DISABLE_UNIT'
  | 'GRANT_TRAIT'
  | 'REMOVE_TRAIT'
  | 'GAIN_MANA'
  | 'REDUCE_COST'

export type TargetType =
  | 'Self'
  | 'EnemyLeader'
  | 'AllyLeader'
  | 'EnemyUnit'
  | 'AllyUnit'
  | 'AllUnits'
  | 'AllEnemyUnits'
  | 'AllAllyUnits'
  | 'RandomEnemyUnit'
  | 'RandomAllyUnit'

export type EffectChainNodeType =
  | 'THEN'
  | 'AND'
  | 'IF_ELSE'
  | 'REPEAT'
  | 'FOREACH'

export interface TargetSelector {
  type: TargetType
  filter?: ConditionFilter
}

export interface ConditionFilter {
  conditionType: string
  parameters: string[]
}

export interface AtomicEffect {
  type: AtomicEffectType
  target?: TargetSelector
  value?: number
  cardId?: string
  trait?: Trait
}

export interface EffectChainNode {
  type: EffectChainNodeType
  atomicEffect?: AtomicEffect
  next?: EffectChainNode
  children?: EffectChainNode[]
  thenNode?: EffectChainNode
  elseNode?: EffectChainNode
  condition?: ConditionFilter
  repeatEffect?: EffectChainNode
  repeatCount?: number
  foreachEffect?: EffectChainNode
  foreachTarget?: TargetSelector
}

export interface EffectDefinition {
  requireTarget: boolean
  root?: EffectChainNode
}

export interface CardEffect {
  cardId: string
  definitions: EffectDefinition[]
}

// バリデーション
export function validateCard(card: Partial<Card>): string[] {
  const errors: string[] = []

  if (!card.id || card.id.trim() === '') {
    errors.push('カードIDは必須です')
  }

  if (!card.name || card.name.trim() === '') {
    errors.push('カード名は必須です')
  }

  if (!card.type) {
    errors.push('カードタイプは必須です')
  }

  if (card.cost === undefined || card.cost < 0) {
    errors.push('コストは0以上である必要があります')
  }

  if (card.type === 'Unit') {
    if (card.attack === undefined || card.defense === undefined) {
      errors.push('ユニットカードには攻撃力と防御力が必要です')
    }
    if (card.attack !== undefined && card.attack < 0) {
      errors.push('攻撃力は0以上である必要があります')
    }
    if (card.defense !== undefined && card.defense < 0) {
      errors.push('防御力は0以上である必要があります')
    }
  }

  if (card.type === 'Spell' || card.type === 'Leader') {
    if (card.traits && card.traits.length > 0) {
      errors.push('スペルおよびリーダーカードには特性を設定できません')
    }
    if (card.attack !== undefined || card.defense !== undefined) {
      errors.push(
        'スペルおよびリーダーカードには攻撃力と防御力を設定できません',
      )
    }
  }

  return errors
}

export function isValidCard(card: Partial<Card>): boolean {
  return validateCard(card).length === 0
}

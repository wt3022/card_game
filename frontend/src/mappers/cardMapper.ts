/**
 * マッパー層: Proto ↔ Domain 変換
 * カードモデルの変換ロジック
 */

import type {
  AtomicEffect,
  AtomicEffectType,
  Card,
  CardEffect,
  CardType,
  ConditionFilter,
  EffectChainNode,
  EffectChainNodeType,
  EffectDefinition,
  TargetSelector,
  TargetType,
  Trait,
} from '../domain/models/Card'
import {
  type AtomicEffect as ProtoAtomicEffect,
  type AtomicEffectType as ProtoAtomicEffectType,
  AtomicEffectType as ProtoAtomicEffectTypeEnum,
  type Card as ProtoCard,
  type CardEffect as ProtoCardEffect,
  type CardType as ProtoCardType,
  CardType as ProtoCardTypeEnum,
  type ConditionFilter as ProtoConditionFilter,
  type EffectChainNode as ProtoEffectChainNode,
  type EffectChainNodeType as ProtoEffectChainNodeType,
  EffectChainNodeType as ProtoEffectChainNodeTypeEnum,
  type EffectDefinition as ProtoEffectDefinition,
  type TargetSelector as ProtoTargetSelector,
  type TargetType as ProtoTargetType,
  TargetType as ProtoTargetTypeEnum,
  type Trait as ProtoTrait,
  Trait as ProtoTraitEnum,
} from '../gen/common_pb'

/**
 * CardMapper: カードモデルのProto ↔ Domain変換
 * すべてのメソッドはstaticで、クラスのインスタンス化は不要
 */
// biome-ignore lint/complexity/noStaticOnlyClass: Mapperクラスは名前空間として使用
export class CardMapper {
  // Proto → Domain
  static toDomain(proto: ProtoCard): Card {
    return {
      id: proto.id,
      instanceId: proto.instanceId || undefined,
      name: proto.name,
      type: CardMapper.cardTypeToDomain(proto.type),
      cost: proto.cost,
      attack: proto.attack ?? undefined,
      defense: proto.defense ?? undefined,
      effect: proto.effect,
      traits: proto.traits.map((t: ProtoTrait) => CardMapper.traitToDomain(t)),
      cardEffect: proto.cardEffect
        ? CardMapper.cardEffectToDomain(proto.cardEffect)
        : undefined,
    }
  }

  // Domain → Proto
  static toProto(domain: Card): Partial<ProtoCard> {
    return {
      id: domain.id,
      instanceId: domain.instanceId || '',
      name: domain.name,
      type: CardMapper.cardTypeToProto(domain.type),
      cost: domain.cost,
      attack: domain.attack,
      defense: domain.defense,
      effect: domain.effect,
      traits: domain.traits.map((t: Trait) => CardMapper.traitToProto(t)),
      cardEffect: domain.cardEffect
        ? (CardMapper.cardEffectToProto(
            domain.cardEffect,
          ) as ProtoCard['cardEffect'])
        : undefined,
    }
  }

  // CardType変換
  private static cardTypeToDomain(proto: ProtoCardType): CardType {
    switch (proto) {
      case ProtoCardTypeEnum.UNIT:
        return 'Unit'
      case ProtoCardTypeEnum.SPELL:
        return 'Spell'
      case ProtoCardTypeEnum.LEADER:
        return 'Leader'
      default:
        return 'Unit'
    }
  }

  private static cardTypeToProto(domain: CardType): ProtoCardType {
    switch (domain) {
      case 'Unit':
        return ProtoCardTypeEnum.UNIT
      case 'Spell':
        return ProtoCardTypeEnum.SPELL
      case 'Leader':
        return ProtoCardTypeEnum.LEADER
    }
  }

  // Trait変換 (gameMapper.tsから使用されるためpublic)
  static traitToDomain(proto: ProtoTrait): Trait {
    switch (proto) {
      case ProtoTraitEnum.RUSH:
        return 'Rush'
      case ProtoTraitEnum.CHARGE:
        return 'Charge'
      case ProtoTraitEnum.WINDFURY:
        return 'Windfury'
      case ProtoTraitEnum.PIERCE:
        return 'Pierce'
      case ProtoTraitEnum.GUARDIAN:
        return 'Guardian'
      case ProtoTraitEnum.EFFECT_SHIELD:
        return 'EffectShield'
      case ProtoTraitEnum.UNTARGETABLE:
        return 'Untargetable'
      default:
        return 'Rush'
    }
  }

  static traitToProto(domain: Trait): ProtoTrait {
    switch (domain) {
      case 'Rush':
        return ProtoTraitEnum.RUSH
      case 'Charge':
        return ProtoTraitEnum.CHARGE
      case 'Windfury':
        return ProtoTraitEnum.WINDFURY
      case 'Pierce':
        return ProtoTraitEnum.PIERCE
      case 'Guardian':
        return ProtoTraitEnum.GUARDIAN
      case 'EffectShield':
        return ProtoTraitEnum.EFFECT_SHIELD
      case 'Untargetable':
        return ProtoTraitEnum.UNTARGETABLE
    }
  }

  // CardEffect変換
  private static cardEffectToDomain(proto: ProtoCardEffect): CardEffect {
    return {
      cardId: proto.cardId,
      definitions: proto.definitions.map((d: ProtoEffectDefinition) =>
        CardMapper.effectDefinitionToDomain(d),
      ),
    }
  }

  private static cardEffectToProto(
    domain: CardEffect,
  ): Partial<ProtoCardEffect> {
    return {
      cardId: domain.cardId,
      definitions: domain.definitions.map((d: EffectDefinition) =>
        CardMapper.effectDefinitionToProto(d),
      ) as ProtoCardEffect['definitions'],
    }
  }

  // EffectDefinition変換
  private static effectDefinitionToDomain(
    proto: ProtoEffectDefinition,
  ): EffectDefinition {
    return {
      requireTarget: proto.requireTarget,
      root: proto.root
        ? CardMapper.effectChainNodeToDomain(proto.root)
        : undefined,
    }
  }

  private static effectDefinitionToProto(
    domain: EffectDefinition,
  ): Partial<ProtoEffectDefinition> {
    return {
      requireTarget: domain.requireTarget,
      root: domain.root
        ? (CardMapper.effectChainNodeToProto(
            domain.root,
          ) as ProtoEffectDefinition['root'])
        : undefined,
    }
  }

  // EffectChainNode変換
  private static effectChainNodeToDomain(
    proto: ProtoEffectChainNode,
  ): EffectChainNode {
    const node: EffectChainNode = {
      type: CardMapper.effectChainNodeTypeToDomain(proto.type),
      atomicEffect: proto.atomicEffect
        ? CardMapper.atomicEffectToDomain(proto.atomicEffect)
        : undefined,
    }

    if (proto.next) node.next = CardMapper.effectChainNodeToDomain(proto.next)
    if (proto.children)
      node.children = proto.children.map((c: ProtoEffectChainNode) =>
        CardMapper.effectChainNodeToDomain(c),
      )
    if (proto.thenNode)
      node.thenNode = CardMapper.effectChainNodeToDomain(proto.thenNode)
    if (proto.elseNode)
      node.elseNode = CardMapper.effectChainNodeToDomain(proto.elseNode)
    if (proto.condition)
      node.condition = CardMapper.conditionFilterToDomain(proto.condition)
    if (proto.repeatEffect)
      node.repeatEffect = CardMapper.effectChainNodeToDomain(proto.repeatEffect)
    if (proto.repeatCount) node.repeatCount = proto.repeatCount
    if (proto.foreachEffect)
      node.foreachEffect = CardMapper.effectChainNodeToDomain(
        proto.foreachEffect,
      )
    if (proto.foreachTarget)
      node.foreachTarget = CardMapper.targetSelectorToDomain(
        proto.foreachTarget,
      )

    return node
  }

  private static effectChainNodeToProto(
    domain: EffectChainNode,
  ): Partial<ProtoEffectChainNode> {
    const proto: Partial<ProtoEffectChainNode> = {
      type: CardMapper.effectChainNodeTypeToProto(domain.type),
      atomicEffect: domain.atomicEffect
        ? (CardMapper.atomicEffectToProto(
            domain.atomicEffect,
          ) as ProtoEffectChainNode['atomicEffect'])
        : undefined,
    }

    if (domain.next)
      proto.next = CardMapper.effectChainNodeToProto(
        domain.next,
      ) as ProtoEffectChainNode['next']
    if (domain.children)
      proto.children = domain.children.map((c: EffectChainNode) =>
        CardMapper.effectChainNodeToProto(c),
      ) as ProtoEffectChainNode['children']
    if (domain.thenNode)
      proto.thenNode = CardMapper.effectChainNodeToProto(
        domain.thenNode,
      ) as ProtoEffectChainNode['thenNode']
    if (domain.elseNode)
      proto.elseNode = CardMapper.effectChainNodeToProto(
        domain.elseNode,
      ) as ProtoEffectChainNode['elseNode']
    if (domain.condition)
      proto.condition = CardMapper.conditionFilterToProto(
        domain.condition,
      ) as ProtoEffectChainNode['condition']
    if (domain.repeatEffect)
      proto.repeatEffect = CardMapper.effectChainNodeToProto(
        domain.repeatEffect,
      ) as ProtoEffectChainNode['repeatEffect']
    if (domain.repeatCount) proto.repeatCount = domain.repeatCount
    if (domain.foreachEffect)
      proto.foreachEffect = CardMapper.effectChainNodeToProto(
        domain.foreachEffect,
      ) as ProtoEffectChainNode['foreachEffect']
    if (domain.foreachTarget)
      proto.foreachTarget = CardMapper.targetSelectorToProto(
        domain.foreachTarget,
      ) as ProtoEffectChainNode['foreachTarget']

    return proto
  }

  // AtomicEffect変換
  private static atomicEffectToDomain(proto: ProtoAtomicEffect): AtomicEffect {
    return {
      type: CardMapper.atomicEffectTypeToDomain(proto.type),
      target: proto.target
        ? CardMapper.targetSelectorToDomain(proto.target)
        : undefined,
      value: proto.value ?? undefined,
      cardId: proto.cardId ?? undefined,
      trait: proto.trait ? CardMapper.traitToDomain(proto.trait) : undefined,
    }
  }

  private static atomicEffectToProto(
    domain: AtomicEffect,
  ): Partial<ProtoAtomicEffect> {
    return {
      type: CardMapper.atomicEffectTypeToProto(domain.type),
      target: domain.target
        ? (CardMapper.targetSelectorToProto(
            domain.target,
          ) as ProtoAtomicEffect['target'])
        : undefined,
      value: domain.value,
      cardId: domain.cardId,
      trait: domain.trait ? CardMapper.traitToProto(domain.trait) : undefined,
    }
  }

  // 型変換ヘルパー
  private static effectChainNodeTypeToDomain(
    proto: ProtoEffectChainNodeType,
  ): EffectChainNodeType {
    switch (proto) {
      case ProtoEffectChainNodeTypeEnum.THEN:
        return 'THEN'
      case ProtoEffectChainNodeTypeEnum.AND:
        return 'AND'
      case ProtoEffectChainNodeTypeEnum.IF_ELSE:
        return 'IF_ELSE'
      case ProtoEffectChainNodeTypeEnum.REPEAT:
        return 'REPEAT'
      case ProtoEffectChainNodeTypeEnum.FOREACH:
        return 'FOREACH'
      default:
        return 'THEN'
    }
  }

  private static effectChainNodeTypeToProto(
    domain: EffectChainNodeType,
  ): ProtoEffectChainNodeType {
    switch (domain) {
      case 'THEN':
        return ProtoEffectChainNodeTypeEnum.THEN
      case 'AND':
        return ProtoEffectChainNodeTypeEnum.AND
      case 'IF_ELSE':
        return ProtoEffectChainNodeTypeEnum.IF_ELSE
      case 'REPEAT':
        return ProtoEffectChainNodeTypeEnum.REPEAT
      case 'FOREACH':
        return ProtoEffectChainNodeTypeEnum.FOREACH
    }
  }

  private static atomicEffectTypeToDomain(
    proto: ProtoAtomicEffectType,
  ): AtomicEffectType {
    const typeMap: Record<number, AtomicEffectType> = {
      [ProtoAtomicEffectTypeEnum.DEAL_DAMAGE]: 'DEAL_DAMAGE',
      [ProtoAtomicEffectTypeEnum.DEAL_SPLASH]: 'DEAL_SPLASH',
      [ProtoAtomicEffectTypeEnum.RESTORE_HP]: 'RESTORE_HP',
      [ProtoAtomicEffectTypeEnum.RESTORE_MANA]: 'RESTORE_MANA',
      [ProtoAtomicEffectTypeEnum.FULL_RESTORE]: 'FULL_RESTORE',
      [ProtoAtomicEffectTypeEnum.DRAW_CARD]: 'DRAW_CARD',
      [ProtoAtomicEffectTypeEnum.DISCARD_CARD]: 'DISCARD_CARD',
      [ProtoAtomicEffectTypeEnum.SEARCH_CARD]: 'SEARCH_CARD',
      [ProtoAtomicEffectTypeEnum.SHUFFLE_DECK]: 'SHUFFLE_DECK',
      [ProtoAtomicEffectTypeEnum.MODIFY_ATTACK]: 'MODIFY_ATTACK',
      [ProtoAtomicEffectTypeEnum.MODIFY_DEFENSE]: 'MODIFY_DEFENSE',
      [ProtoAtomicEffectTypeEnum.MODIFY_COST]: 'MODIFY_COST',
      [ProtoAtomicEffectTypeEnum.MODIFY_MAX_HP]: 'MODIFY_MAX_HP',
      [ProtoAtomicEffectTypeEnum.SUMMON_UNIT]: 'SUMMON_UNIT',
      [ProtoAtomicEffectTypeEnum.DESTROY_UNIT]: 'DESTROY_UNIT',
      [ProtoAtomicEffectTypeEnum.RETURN_TO_HAND]: 'RETURN_TO_HAND',
      [ProtoAtomicEffectTypeEnum.RETURN_TO_DECK]: 'RETURN_TO_DECK',
      [ProtoAtomicEffectTypeEnum.DISABLE_UNIT]: 'DISABLE_UNIT',
      [ProtoAtomicEffectTypeEnum.GRANT_TRAIT]: 'GRANT_TRAIT',
      [ProtoAtomicEffectTypeEnum.REMOVE_TRAIT]: 'REMOVE_TRAIT',
      [ProtoAtomicEffectTypeEnum.GAIN_MANA]: 'GAIN_MANA',
      [ProtoAtomicEffectTypeEnum.REDUCE_COST]: 'REDUCE_COST',
    }
    return typeMap[proto] || 'DEAL_DAMAGE'
  }

  private static atomicEffectTypeToProto(
    domain: AtomicEffectType,
  ): ProtoAtomicEffectType {
    const typeMap: Record<AtomicEffectType, ProtoAtomicEffectType> = {
      DEAL_DAMAGE: ProtoAtomicEffectTypeEnum.DEAL_DAMAGE,
      DEAL_SPLASH: ProtoAtomicEffectTypeEnum.DEAL_SPLASH,
      RESTORE_HP: ProtoAtomicEffectTypeEnum.RESTORE_HP,
      RESTORE_MANA: ProtoAtomicEffectTypeEnum.RESTORE_MANA,
      FULL_RESTORE: ProtoAtomicEffectTypeEnum.FULL_RESTORE,
      DRAW_CARD: ProtoAtomicEffectTypeEnum.DRAW_CARD,
      DISCARD_CARD: ProtoAtomicEffectTypeEnum.DISCARD_CARD,
      SEARCH_CARD: ProtoAtomicEffectTypeEnum.SEARCH_CARD,
      SHUFFLE_DECK: ProtoAtomicEffectTypeEnum.SHUFFLE_DECK,
      MODIFY_ATTACK: ProtoAtomicEffectTypeEnum.MODIFY_ATTACK,
      MODIFY_DEFENSE: ProtoAtomicEffectTypeEnum.MODIFY_DEFENSE,
      MODIFY_COST: ProtoAtomicEffectTypeEnum.MODIFY_COST,
      MODIFY_MAX_HP: ProtoAtomicEffectTypeEnum.MODIFY_MAX_HP,
      SUMMON_UNIT: ProtoAtomicEffectTypeEnum.SUMMON_UNIT,
      DESTROY_UNIT: ProtoAtomicEffectTypeEnum.DESTROY_UNIT,
      RETURN_TO_HAND: ProtoAtomicEffectTypeEnum.RETURN_TO_HAND,
      RETURN_TO_DECK: ProtoAtomicEffectTypeEnum.RETURN_TO_DECK,
      DISABLE_UNIT: ProtoAtomicEffectTypeEnum.DISABLE_UNIT,
      GRANT_TRAIT: ProtoAtomicEffectTypeEnum.GRANT_TRAIT,
      REMOVE_TRAIT: ProtoAtomicEffectTypeEnum.REMOVE_TRAIT,
      GAIN_MANA: ProtoAtomicEffectTypeEnum.GAIN_MANA,
      REDUCE_COST: ProtoAtomicEffectTypeEnum.REDUCE_COST,
    }
    return typeMap[domain]
  }

  private static targetTypeToDomain(proto: ProtoTargetType): TargetType {
    const typeMap: Record<number, TargetType> = {
      [ProtoTargetTypeEnum.SELF]: 'Self',
      [ProtoTargetTypeEnum.ENEMY_LEADER]: 'EnemyLeader',
      [ProtoTargetTypeEnum.ALLY_LEADER]: 'AllyLeader',
      [ProtoTargetTypeEnum.ENEMY_UNIT]: 'EnemyUnit',
      [ProtoTargetTypeEnum.ALLY_UNIT]: 'AllyUnit',
      [ProtoTargetTypeEnum.ALL_UNITS]: 'AllUnits',
      [ProtoTargetTypeEnum.ALL_ENEMY_UNITS]: 'AllEnemyUnits',
      [ProtoTargetTypeEnum.ALL_ALLY_UNITS]: 'AllAllyUnits',
      [ProtoTargetTypeEnum.RANDOM_ENEMY_UNIT]: 'RandomEnemyUnit',
      [ProtoTargetTypeEnum.RANDOM_ALLY_UNIT]: 'RandomAllyUnit',
    }
    return typeMap[proto] || 'Self'
  }

  private static targetTypeToProto(domain: TargetType): ProtoTargetType {
    const typeMap: Record<TargetType, ProtoTargetType> = {
      Self: ProtoTargetTypeEnum.SELF,
      EnemyLeader: ProtoTargetTypeEnum.ENEMY_LEADER,
      AllyLeader: ProtoTargetTypeEnum.ALLY_LEADER,
      EnemyUnit: ProtoTargetTypeEnum.ENEMY_UNIT,
      AllyUnit: ProtoTargetTypeEnum.ALLY_UNIT,
      AllUnits: ProtoTargetTypeEnum.ALL_UNITS,
      AllEnemyUnits: ProtoTargetTypeEnum.ALL_ENEMY_UNITS,
      AllAllyUnits: ProtoTargetTypeEnum.ALL_ALLY_UNITS,
      RandomEnemyUnit: ProtoTargetTypeEnum.RANDOM_ENEMY_UNIT,
      RandomAllyUnit: ProtoTargetTypeEnum.RANDOM_ALLY_UNIT,
    }
    return typeMap[domain]
  }

  private static targetSelectorToDomain(
    proto: ProtoTargetSelector,
  ): TargetSelector {
    return {
      type: CardMapper.targetTypeToDomain(proto.type),
      filter: proto.filter
        ? CardMapper.conditionFilterToDomain(proto.filter)
        : undefined,
    }
  }

  private static targetSelectorToProto(
    domain: TargetSelector,
  ): Partial<ProtoTargetSelector> {
    return {
      type: CardMapper.targetTypeToProto(domain.type),
      filter: domain.filter
        ? (CardMapper.conditionFilterToProto(
            domain.filter,
          ) as ProtoTargetSelector['filter'])
        : undefined,
    }
  }

  private static conditionFilterToDomain(
    proto: ProtoConditionFilter,
  ): ConditionFilter {
    return {
      conditionType: proto.conditionType,
      parameters: proto.parameters,
    }
  }

  private static conditionFilterToProto(
    domain: ConditionFilter,
  ): Partial<ProtoConditionFilter> {
    return {
      conditionType: domain.conditionType,
      parameters: domain.parameters,
    }
  }
}

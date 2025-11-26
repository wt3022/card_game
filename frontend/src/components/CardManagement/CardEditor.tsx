import { useCallback, useEffect, useState } from 'react'
import type { Card, Trait } from '../../gen/common_pb'
import {
  AtomicEffectType,
  CardEffect,
  CardType,
  EffectChainNodeType,
  EffectDefinition,
  AtomicEffect as ProtoAtomicEffect,
  EffectChainNode as ProtoEffectChainNode,
  TargetSelector as ProtoTargetSelector,
  TargetType,
  Trait as TraitEnum,
} from '../../gen/common_pb'
import { cardManagementClient } from '../../lib/api-client'
import EffectNodeEditor from './EffectNodeEditor'
import './CardEditor.css'

// EffectChainNodeの型定義をエクスポート
export type EffectNodeType = 'THEN' | 'AND' | 'IF_ELSE' | 'REPEAT' | 'FOREACH'

export interface TargetFilter {
  min_attack?: number | null
  max_attack?: number | null
  min_defense?: number | null
  max_defense?: number | null
  min_cost?: number | null
  max_cost?: number | null
  card_type?: string | null
  has_traits?: Array<{ trait: string }> | null
  lack_traits?: Array<{ trait: string }> | null
}

export interface Condition {
  type: string
  operator: string
  value: number
}

export interface TargetSelector {
  type: string
  count: number
  random: boolean
  select_by_max: string
  select_by_min: string
  filter: TargetFilter | null
}

export interface AtomicEffect {
  type: string
  value: number
  multiplier: number
  duration: number | null
  timing: string
  target: TargetSelector
  condition: Condition | null
  parameters: string
}

export interface EffectChainNode {
  type: EffectNodeType
  atomic_effect?: AtomicEffect | null
  sequential?: {
    next_id: number | null
    next: EffectChainNode | null
  }
  parallel?: {
    children: EffectChainNode[]
    parallel_next: EffectChainNode | null
  }
  if_else?: {
    condition: Condition
    thenNode: EffectChainNode
    elseNode: EffectChainNode | null
  }
  repeat?: {
    repeat_effect: EffectChainNode
    count: number
  }
  for_each?: {
    for_each_effect: EffectChainNode
    for_each_target: TargetSelector
  }
}

type BackendTargetSelector = Partial<TargetSelector> & {
  filter?: TargetFilter | null
}

type BackendAtomicEffect = {
  type?: string
  value?: number
  multiplier?: number
  duration?: number | null
  timing?: string
  target?: BackendTargetSelector | null
  condition?: Condition | null
  parameters?: Record<string, unknown> | string | null
}

type BackendSequentialNode = {
  effect?: BackendAtomicEffect | null
  next?: BackendEffectNode | null
  next_id?: number | null
}

type BackendParallelNode = {
  children?: BackendEffectNode[]
  parallel_next?: BackendEffectNode | null
}

type BackendIfElseNode = {
  condition?: Condition
  then?: BackendEffectNode
  else?: BackendEffectNode | null
}

type BackendRepeatNode = {
  repeat_effect?: BackendEffectNode
  count?: number
}

type BackendForEachNode = {
  for_each_effect?: BackendEffectNode
  for_each_target?: BackendTargetSelector | null
}

type BackendEffectNode = {
  type: EffectNodeType
  atomic_effect?: BackendAtomicEffect | null
  sequential?: BackendSequentialNode | null
  parallel?: BackendParallelNode | null
  if_else?: BackendIfElseNode | null
  repeat?: BackendRepeatNode | null
  for_each?: BackendForEachNode | null
}

// TODO: 効果の保存機能実装時に使用
// interface BackendCardEffect {
//   definitions?: Array<{
//     root?: BackendEffectNode | null
//   }>
// }

// 特性名の日本語マッピング
const traitNames: Record<TraitEnum, string> = {
  [TraitEnum.UNSPECIFIED]: '未指定',
  [TraitEnum.RUSH]: '疾走',
  [TraitEnum.CHARGE]: '突進',
  [TraitEnum.WINDFURY]: '疾風',
  [TraitEnum.PIERCE]: '貫通',
  [TraitEnum.GUARDIAN]: '守護',
  [TraitEnum.EFFECT_SHIELD]: '効果盾',
  [TraitEnum.UNTARGETABLE]: '対象不可',
}

// デフォルトの原子効果を作成（エクスポート）
export const createDefaultAtomicEffect = (): AtomicEffect => ({
  type: 'DEAL_DAMAGE',
  value: 0,
  multiplier: 1.0,
  duration: null,
  timing: 'Immediate',
  target: {
    type: 'Self',
    count: 0,
    random: false,
    select_by_max: '',
    select_by_min: '',
    filter: null,
  } as TargetSelector,
  condition: null,
  parameters: '{}',
})

// デフォルトのEffectChainNodeを作成（エクスポート）
export const createDefaultEffectNode = (
  type: EffectNodeType = 'THEN',
): EffectChainNode => {
  switch (type) {
    case 'THEN':
      return {
        type: 'THEN',
        atomic_effect: createDefaultAtomicEffect(),
        sequential: {
          next_id: null,
          next: null,
        },
      }
    case 'AND':
      return {
        type: 'AND',
        parallel: {
          children: [],
          parallel_next: null,
        },
      }
    case 'IF_ELSE':
      return {
        type: 'IF_ELSE',
        if_else: {
          condition: {
            type: 'PLAYER_HP',
            operator: 'GREATER_THAN',
            value: 0,
          },
          thenNode: createDefaultEffectNode('THEN'),
          elseNode: null,
        },
      }
    case 'REPEAT':
      return {
        type: 'REPEAT',
        repeat: {
          repeat_effect: createDefaultEffectNode('THEN'),
          count: 1,
        },
      }
    case 'FOREACH':
      return {
        type: 'FOREACH',
        for_each: {
          for_each_effect: createDefaultEffectNode('THEN'),
          for_each_target: {
            type: 'Allies',
            count: 0,
            random: false,
            select_by_max: '',
            select_by_min: '',
            filter: null,
          } as TargetSelector,
        },
      }
    default:
      return createDefaultEffectNode('THEN')
  }
}

interface CardEditorProps {
  card: Card | null
  isNewCardMode?: boolean
  onSave: (cardId: string) => void
  onCancel: () => void
}

export default function CardEditor({
  card,
  isNewCardMode = false,
  onSave,
  onCancel,
}: CardEditorProps) {
  const [formData, setFormData] = useState({
    id: '',
    name: '',
    type: CardType.UNIT,
    cost: 0,
    attack: undefined as number | undefined,
    defense: undefined as number | undefined,
    effectText: '',
    traits: [] as Trait[],
    // カード効果
    hasEffect: false,
    effectRoot: null as EffectChainNode | null,
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // protobuf TargetSelector を UI TargetSelector に変換
  const convertProtoTargetToUI = useCallback(
    (protoTarget?: ProtoTargetSelector): TargetSelector => {
      if (!protoTarget) {
        return {
          type: 'Self',
          count: 0,
          random: false,
          select_by_max: '',
          select_by_min: '',
          filter: null,
        }
      }
      // TODO: filter の変換を実装
      return {
        type: TargetType[protoTarget.type] || 'Self',
        count: 0,
        random: false,
        select_by_max: '',
        select_by_min: '',
        filter: null,
      }
    },
    [],
  )

  // protobuf AtomicEffect を UI AtomicEffect に変換
  const convertProtoAtomicToUI = useCallback(
    (protoEffect?: ProtoAtomicEffect): AtomicEffect => {
      if (!protoEffect) {
        return createDefaultAtomicEffect()
      }
      return {
        type: AtomicEffectType[protoEffect.type] || 'DEAL_DAMAGE',
        value: protoEffect.value || 0,
        multiplier: 1.0,
        duration: null,
        timing: 'Immediate',
        target: convertProtoTargetToUI(protoEffect.target),
        condition: null,
        parameters: '{}',
      }
    },
    [convertProtoTargetToUI],
  )

  // protobuf EffectChainNode を UI EffectChainNode に変換
  const convertProtoNodeToUI = useCallback(
    (protoNode?: ProtoEffectChainNode): EffectChainNode | null => {
      if (!protoNode) return null

      const nodeType = EffectChainNodeType[protoNode.type] as EffectNodeType

      switch (nodeType) {
        case 'THEN':
          return {
            type: 'THEN',
            atomic_effect: convertProtoAtomicToUI(protoNode.atomicEffect),
            sequential: {
              next_id: null,
              next: convertProtoNodeToUI(protoNode.next),
            },
          }
        case 'AND':
          return {
            type: 'AND',
            parallel: {
              children:
                protoNode.children
                  ?.map((child: ProtoEffectChainNode) =>
                    convertProtoNodeToUI(child),
                  )
                  .filter((node): node is EffectChainNode => node !== null) ||
                [],
              parallel_next: null,
            },
          }
        case 'IF_ELSE':
          // TODO: condition の変換を実装
          return {
            type: 'IF_ELSE',
            if_else: {
              condition: {
                type: 'PLAYER_HP',
                operator: 'GREATER_THAN',
                value: 0,
              },
              thenNode:
                convertProtoNodeToUI(protoNode.thenNode) ||
                createDefaultEffectNode('THEN'),
              elseNode: convertProtoNodeToUI(protoNode.elseNode),
            },
          }
        case 'REPEAT':
          return {
            type: 'REPEAT',
            repeat: {
              repeat_effect:
                convertProtoNodeToUI(protoNode.repeatEffect) ||
                createDefaultEffectNode('THEN'),
              count: protoNode.repeatCount || 1,
            },
          }
        case 'FOREACH':
          return {
            type: 'FOREACH',
            for_each: {
              for_each_effect:
                convertProtoNodeToUI(protoNode.foreachEffect) ||
                createDefaultEffectNode('THEN'),
              for_each_target: convertProtoTargetToUI(protoNode.foreachTarget),
            },
          }
        default:
          return null
      }
    },
    [convertProtoAtomicToUI, convertProtoTargetToUI],
  )

  const normalizeTargetSelector = useCallback(
    (selector?: BackendTargetSelector | null): TargetSelector => ({
      type: selector?.type ?? 'Self',
      count: selector?.count ?? 0,
      random: selector?.random ?? false,
      select_by_max: selector?.select_by_max ?? '',
      select_by_min: selector?.select_by_min ?? '',
      filter: selector?.filter ?? null,
    }),
    [],
  )

  const normalizeAtomicEffect = useCallback(
    (effect?: BackendAtomicEffect | null): AtomicEffect => {
      const safeEffect: BackendAtomicEffect = effect ?? {}
      const rawParameters = safeEffect.parameters
      let parameters = '{}'
      if (typeof rawParameters === 'string') {
        parameters = rawParameters || '{}'
      } else if (rawParameters && typeof rawParameters === 'object') {
        try {
          parameters = JSON.stringify(rawParameters)
        } catch {
          parameters = '{}'
        }
      }

      return {
        type: safeEffect.type ?? 'DEAL_DAMAGE',
        value: typeof safeEffect.value === 'number' ? safeEffect.value : 0,
        multiplier:
          typeof safeEffect.multiplier === 'number' ? safeEffect.multiplier : 1,
        duration:
          typeof safeEffect.duration === 'number' ||
          safeEffect.duration === null
            ? safeEffect.duration
            : null,
        timing: safeEffect.timing ?? 'Immediate',
        target: normalizeTargetSelector(safeEffect.target ?? null),
        condition: safeEffect.condition ?? null,
        parameters,
      }
    },
    [normalizeTargetSelector],
  )

  // フロントエンドの型(thenNode/elseNode)をバックエンドの型(then/else)に変換
  const _convertNodeForBackend = (
    node: EffectChainNode,
  ): Record<string, unknown> => {
    const converted: Record<string, unknown> = {
      type: node.type,
    }

    if (node.atomic_effect) {
      converted.atomic_effect = node.atomic_effect
    }

    if (node.sequential) {
      converted.sequential = {
        next_id: node.sequential.next_id,
        next: node.sequential.next
          ? _convertNodeForBackend(node.sequential.next)
          : null,
      }
    }

    if (node.parallel) {
      converted.parallel = {
        children: node.parallel.children.map((child) =>
          _convertNodeForBackend(child),
        ),
        parallel_next: node.parallel.parallel_next
          ? _convertNodeForBackend(node.parallel.parallel_next)
          : null,
      }
    }

    if (node.if_else) {
      converted.if_else = {
        condition: node.if_else.condition,
        // biome-ignore lint/suspicious/noThenProperty: バックエンドのJSON構造に合わせる必要があるため
        then: _convertNodeForBackend(node.if_else.thenNode),
        else: node.if_else.elseNode
          ? _convertNodeForBackend(node.if_else.elseNode)
          : null,
      }
    }

    if (node.repeat) {
      converted.repeat = {
        repeat_effect: _convertNodeForBackend(node.repeat.repeat_effect),
        count: node.repeat.count,
      }
    }

    if (node.for_each) {
      converted.for_each = {
        for_each_effect: _convertNodeForBackend(node.for_each.for_each_effect),
        for_each_target: node.for_each.for_each_target,
      }
    }

    return converted
  }

  // バックエンドJSON → UIノード変換
  const convertNodeFromBackend = useCallback(
    (node: BackendEffectNode): EffectChainNode => {
      const converted: EffectChainNode = {
        type: node.type,
      }

      const backendAtomic =
        node.atomic_effect ?? node.sequential?.effect ?? null
      if (backendAtomic) {
        converted.atomic_effect = normalizeAtomicEffect(backendAtomic)
      }

      if (node.sequential) {
        converted.sequential = {
          next_id: node.sequential.next_id ?? null,
          next: node.sequential.next
            ? convertNodeFromBackend(node.sequential.next)
            : null,
        }
      }

      if (node.parallel) {
        const children =
          node.parallel.children?.filter((child): child is BackendEffectNode =>
            Boolean(child),
          ) ?? []
        converted.parallel = {
          children: children.map((child) => convertNodeFromBackend(child)),
          parallel_next: node.parallel.parallel_next
            ? convertNodeFromBackend(node.parallel.parallel_next)
            : null,
        }
      }

      if (node.if_else?.then) {
        const condition: Condition = node.if_else.condition ?? {
          type: 'PLAYER_HP',
          operator: 'GREATER_THAN',
          value: 0,
        }
        converted.if_else = {
          condition,
          thenNode: convertNodeFromBackend(node.if_else.then),
          elseNode: node.if_else.else
            ? convertNodeFromBackend(node.if_else.else)
            : null,
        }
      }

      if (node.repeat?.repeat_effect) {
        converted.repeat = {
          repeat_effect: convertNodeFromBackend(node.repeat.repeat_effect),
          count: node.repeat.count ?? 1,
        }
      }

      if (node.for_each?.for_each_effect) {
        converted.for_each = {
          for_each_effect: convertNodeFromBackend(
            node.for_each.for_each_effect,
          ),
          for_each_target: normalizeTargetSelector(
            node.for_each.for_each_target ?? null,
          ),
        }
      }

      return converted
    },
    [normalizeAtomicEffect, normalizeTargetSelector],
  )

  useEffect(() => {
    if (card) {
      let hasEffect = false
      let effectRoot: EffectChainNode | null = null

      // cardEffectからエフェクトデータを読み込む
      if (card.cardEffect) {
        try {
          // CardEffectオブジェクトからエフェクトノードを変換
          const definition = card.cardEffect.definitions?.[0]
          if (definition?.root) {
            hasEffect = true
            effectRoot = convertProtoNodeToUI(definition.root)
          }
        } catch (err) {
          console.error('カード効果のパースに失敗', err)
        }
      }

      setFormData({
        id: card.id,
        name: card.name,
        type: card.type,
        cost: card.cost,
        attack: card.attack,
        defense: card.defense,
        effectText: card.effect,
        traits: card.traits || [],
        hasEffect,
        effectRoot,
      })
    } else if (isNewCardMode) {
      setFormData({
        id: '',
        name: '',
        type: CardType.UNIT,
        cost: 0,
        attack: undefined,
        defense: undefined,
        effectText: '',
        traits: [],
        hasEffect: false,
        effectRoot: null,
      })
    }
  }, [card, isNewCardMode, convertProtoNodeToUI])

  // EffectChainNode → ProtoEffectChainNode変換
  const convertNodeToProto = useCallback(
    (node: EffectChainNode): ProtoEffectChainNode | undefined => {
      // ノードタイプをproto enumに変換
      const getProtoNodeType = (type: EffectNodeType): EffectChainNodeType => {
        const typeMap: Record<EffectNodeType, EffectChainNodeType> = {
          THEN: EffectChainNodeType.THEN,
          AND: EffectChainNodeType.AND,
          IF_ELSE: EffectChainNodeType.IF_ELSE,
          REPEAT: EffectChainNodeType.REPEAT,
          FOREACH: EffectChainNodeType.FOREACH,
        }
        return typeMap[type] || EffectChainNodeType.UNSPECIFIED
      }

      const protoNode = new ProtoEffectChainNode({
        id: 0, // バックエンドで自動生成
        type: getProtoNodeType(node.type),
      })

      // AtomicEffectの変換
      if (node.atomic_effect) {
        // 文字列をenumに変換(簡易実装:完全な変換は後で実装)
        const effectType = (
          AtomicEffectType as unknown as Record<string, AtomicEffectType>
        )[node.atomic_effect.type] as AtomicEffectType | undefined

        const targetType = node.atomic_effect.target
          ? ((TargetType as unknown as Record<string, TargetType>)[
              node.atomic_effect.target.type
            ] as TargetType | undefined)
          : undefined

        protoNode.atomicEffect = new ProtoAtomicEffect({
          id: 0, // バックエンドで自動生成
          type: effectType ?? AtomicEffectType.UNSPECIFIED,
          value: node.atomic_effect.value,
          target: node.atomic_effect.target
            ? new ProtoTargetSelector({
                id: 0,
                type: targetType ?? TargetType.UNSPECIFIED,
                // maxCount/minCountなど、protoで定義されているフィールドのみ使用
                // filter: TODO 必要に応じて実装
              })
            : undefined,
        })
      }

      // THENノードの場合
      if (node.sequential?.next) {
        protoNode.next = convertNodeToProto(node.sequential.next)
      }

      // ANDノードの場合
      if (node.parallel?.children) {
        protoNode.children = node.parallel.children
          .map((child) => convertNodeToProto(child))
          .filter((child): child is ProtoEffectChainNode => child !== undefined)
      }

      // IF_ELSEノードの場合
      if (node.if_else) {
        protoNode.thenNode = convertNodeToProto(node.if_else.thenNode)
        if (node.if_else.elseNode) {
          protoNode.elseNode = convertNodeToProto(node.if_else.elseNode)
        }
        // condition: TODO 必要に応じて実装
      }

      // REPEATノードの場合
      if (node.repeat) {
        protoNode.repeatEffect = convertNodeToProto(node.repeat.repeat_effect)
        protoNode.repeatCount = node.repeat.count
      }

      // FOREACHノードの場合
      if (node.for_each) {
        protoNode.foreachEffect = convertNodeToProto(
          node.for_each.for_each_effect,
        )
        // foreachTarget: TODO 必要に応じて実装
      }

      return protoNode
    },
    [],
  )

  // CardEffect protobuf型への変換
  const buildCardEffect = useCallback((): CardEffect | undefined => {
    if (!formData.hasEffect || !formData.effectRoot) {
      return undefined
    }

    const protoRoot = convertNodeToProto(formData.effectRoot)
    if (!protoRoot) {
      return undefined
    }

    return new CardEffect({
      id: 0, // バックエンドで自動生成
      cardId: formData.id,
      definitions: [
        new EffectDefinition({
          id: 0, // バックエンドで自動生成
          requireTarget: false, // TODO: UIから設定できるようにする
          root: protoRoot,
        }),
      ],
    })
  }, [formData.hasEffect, formData.effectRoot, formData.id, convertNodeToProto])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      // CardEffect protobuf型に変換
      const cardEffect = buildCardEffect()
      let savedCardId: string

      if (card) {
        // 更新
        const response = await cardManagementClient.updateCard({
          id: formData.id,
          name: formData.name,
          type: formData.type,
          cost: formData.cost,
          attack: formData.attack,
          defense: formData.defense,
          effectText: '', // 効果テキストは自動生成されるため空文字列を送信
          traits: formData.traits,
          cardEffect,
        })
        savedCardId = response.card?.id || formData.id
      } else {
        // 作成
        const response = await cardManagementClient.createCard({
          id: formData.id,
          name: formData.name,
          type: formData.type,
          cost: formData.cost,
          attack: formData.attack,
          defense: formData.defense,
          effectText: '', // 効果テキストは自動生成されるため空文字列を送信
          traits: formData.traits,
          cardEffect,
        })
        savedCardId = response.card?.id || formData.id
      }

      onSave(savedCardId)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setLoading(false)
    }
  }

  // 初期状態（カードが選択されていない、かつ新規作成モードでもない）
  if (!card && !isNewCardMode) {
    return (
      <div className="card-editor-empty">
        <p>
          左側のリストからカードを選択するか、「新しいカード」をクリックしてください
        </p>
      </div>
    )
  }

  return (
    <div className="card-editor">
      <h2>{card ? 'カード編集' : '新しいカード'}</h2>
      {error && <div className="error-message">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="id">カードID</label>
          <input
            id="id"
            type="text"
            value={formData.id}
            onChange={(e) => setFormData({ ...formData, id: e.target.value })}
            required
            disabled={!!card}
          />
        </div>
        <div className="form-group">
          <label htmlFor="name">カード名</label>
          <input
            id="name"
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="type">タイプ</label>
          <select
            id="type"
            value={formData.type}
            onChange={(e) => {
              const newType = Number(e.target.value) as CardType
              // スペルまたはリーダーの場合は特性をクリア
              const newTraits = newType === CardType.UNIT ? formData.traits : []
              setFormData({
                ...formData,
                type: newType,
                traits: newTraits,
              })
            }}
            required
          >
            <option value={CardType.UNIT}>ユニット</option>
            <option value={CardType.SPELL}>スペル</option>
            <option value={CardType.LEADER}>リーダー</option>
          </select>
        </div>
        <div className="form-group">
          <label htmlFor="cost">コスト</label>
          <input
            id="cost"
            type="number"
            min="0"
            value={formData.cost}
            onChange={(e) =>
              setFormData({ ...formData, cost: Number(e.target.value) })
            }
            required
          />
        </div>
        {formData.type === CardType.UNIT && (
          <>
            <div className="form-group">
              <label htmlFor="attack">攻撃力</label>
              <input
                id="attack"
                type="number"
                min="0"
                value={formData.attack || ''}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    attack: e.target.value ? Number(e.target.value) : undefined,
                  })
                }
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="defense">防御力</label>
              <input
                id="defense"
                type="number"
                min="0"
                value={formData.defense || ''}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    defense: e.target.value
                      ? Number(e.target.value)
                      : undefined,
                  })
                }
                required
              />
            </div>
          </>
        )}
        {formData.type === CardType.UNIT && (
          <div className="form-group">
            <div className="form-label">特性</div>
            <div className="traits-container">
              {Object.values(TraitEnum)
                .filter(
                  (trait) =>
                    typeof trait === 'number' &&
                    trait !== TraitEnum.UNSPECIFIED,
                )
                .map((trait) => {
                  const traitValue = trait as Trait
                  const traitName =
                    traitNames[traitValue] || TraitEnum[traitValue]
                  const isChecked = formData.traits.includes(traitValue)
                  return (
                    <label key={traitValue} className="trait-checkbox">
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setFormData({
                              ...formData,
                              traits: [...formData.traits, traitValue],
                            })
                          } else {
                            setFormData({
                              ...formData,
                              traits: formData.traits.filter(
                                (t) => t !== traitValue,
                              ),
                            })
                          }
                        }}
                      />
                      <span>{traitName}</span>
                    </label>
                  )
                })}
            </div>
          </div>
        )}
        <div className="form-group">
          <label>
            <input
              type="checkbox"
              checked={formData.hasEffect}
              onChange={(e) =>
                setFormData({ ...formData, hasEffect: e.target.checked })
              }
            />
            カード効果を設定する
          </label>
        </div>
        {formData.hasEffect && (
          <div className="card-effect-section">
            <h3>カード効果</h3>
            {!formData.effectRoot && (
              <div className="form-group">
                <button
                  type="button"
                  onClick={() => {
                    setFormData({
                      ...formData,
                      effectRoot: createDefaultEffectNode('THEN'),
                    })
                  }}
                >
                  効果を追加
                </button>
              </div>
            )}
            {formData.effectRoot && (
              <EffectNodeEditor
                node={formData.effectRoot}
                onChange={(newNode) => {
                  setFormData({ ...formData, effectRoot: newNode })
                }}
                onDelete={() => {
                  setFormData({
                    ...formData,
                    effectRoot: null,
                    hasEffect: false,
                  })
                }}
              />
            )}
          </div>
        )}
        <div className="form-actions">
          <button type="submit" disabled={loading}>
            {loading ? '保存中...' : '保存'}
          </button>
          <button type="button" onClick={onCancel} disabled={loading}>
            キャンセル
          </button>
        </div>
      </form>
    </div>
  )
}

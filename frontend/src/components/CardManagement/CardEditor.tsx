import { useCallback, useEffect, useState } from 'react'
import { CardManagementService } from '../../gen/card_management_connect'
import type { Card, Trait } from '../../gen/common_pb'
import { CardType, Trait as TraitEnum } from '../../gen/common_pb'
import { createAuthenticatedClient } from '../../lib/auth'
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

interface BackendCardEffect {
  definitions?: Array<{
    root?: BackendEffectNode | null
  }>
}

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
  onSave: () => void
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

  const cardClient = createAuthenticatedClient(CardManagementService)

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

  // フロントエンドの型（thenNode/elseNode）をバックエンドの型（then/else）に変換
  const convertNodeForBackend = (
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
          ? convertNodeForBackend(node.sequential.next)
          : null,
      }
    }

    if (node.parallel) {
      converted.parallel = {
        children: node.parallel.children.map((child) =>
          convertNodeForBackend(child),
        ),
        parallel_next: node.parallel.parallel_next
          ? convertNodeForBackend(node.parallel.parallel_next)
          : null,
      }
    }

    if (node.if_else) {
      converted.if_else = {
        condition: node.if_else.condition,
        // biome-ignore lint/suspicious/noThenProperty: バックエンドのJSON構造に合わせる必要があるため
        then: convertNodeForBackend(node.if_else.thenNode),
        else: node.if_else.elseNode
          ? convertNodeForBackend(node.if_else.elseNode)
          : null,
      }
    }

    if (node.repeat) {
      converted.repeat = {
        repeat_effect: convertNodeForBackend(node.repeat.repeat_effect),
        count: node.repeat.count,
      }
    }

    if (node.for_each) {
      converted.for_each = {
        for_each_effect: convertNodeForBackend(node.for_each.for_each_effect),
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

      if (card.cardEffectJson) {
        try {
          const parsed = JSON.parse(card.cardEffectJson) as BackendCardEffect
          const definition = parsed?.definitions?.[0]
          if (definition?.root) {
            hasEffect = true
            effectRoot = convertNodeFromBackend(definition.root)
          }
        } catch (err) {
          console.error('Failed to parse card effect json', err)
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
  }, [card, isNewCardMode, convertNodeFromBackend])

  const buildCardEffectJson = (): string => {
    if (!formData.hasEffect || !formData.effectRoot) {
      return ''
    }

    // model構造に合わせたCardEffectModelを構築
    const cardEffectModel = {
      definitions: [
        {
          require_target: false, // TODO: ルートノードから判定
          root: convertNodeForBackend(formData.effectRoot),
        },
      ],
    }

    return JSON.stringify(cardEffectModel)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      const cardEffectJson = buildCardEffectJson()
      if (card) {
        // 更新
        await cardClient.updateCard({
          id: formData.id,
          name: formData.name,
          type: formData.type,
          cost: formData.cost,
          attack: formData.attack,
          defense: formData.defense,
          effectText: '', // 効果テキストは自動生成されるため空文字列を送信
          traits: formData.traits,
          cardEffectJson,
        })
      } else {
        // 作成
        await cardClient.createCard({
          id: formData.id,
          name: formData.name,
          type: formData.type,
          cost: formData.cost,
          attack: formData.attack,
          defense: formData.defense,
          effectText: '', // 効果テキストは自動生成されるため空文字列を送信
          traits: formData.traits,
          cardEffectJson,
        })
      }
      onSave()
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

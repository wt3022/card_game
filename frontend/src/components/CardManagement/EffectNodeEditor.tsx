import type {
  AtomicEffect,
  EffectChainNode,
  EffectNodeType,
  TargetSelector,
} from './CardEditor'
import {
  createDefaultAtomicEffect,
  createDefaultEffectNode,
} from './CardEditor'
import './CardEditor.css'

interface EffectNodeEditorProps {
  node: EffectChainNode
  onChange: (node: EffectChainNode) => void
  onDelete?: () => void
  depth?: number
}

export default function EffectNodeEditor({
  node,
  onChange,
  onDelete,
  depth = 0,
}: EffectNodeEditorProps) {
  const updateNodeType = (newType: EffectNodeType) => {
    // ノードタイプが変更された場合、新しいデフォルトノードを作成
    let newNode: EffectChainNode

    switch (newType) {
      case 'THEN':
        newNode = {
          type: 'THEN',
          atomic_effect: node.atomic_effect || createDefaultAtomicEffect(),
          sequential: {
            next_id: null,
            next: null,
          },
        }
        break
      case 'AND':
        newNode = {
          type: 'AND',
          parallel: {
            children: [],
            parallel_next: null,
          },
        }
        break
      case 'IF_ELSE':
        newNode = {
          type: 'IF_ELSE',
          if_else: {
            condition: {
              type: 'PLAYER_HP',
              operator: 'GREATER_THAN',
              value: 0,
            },
            thenNode: node.atomic_effect
              ? {
                  type: 'THEN',
                  atomic_effect: node.atomic_effect,
                  sequential: { next_id: null, next: null },
                }
              : createDefaultEffectNode('THEN'),
            elseNode: null,
          },
        }
        break
      case 'REPEAT':
        newNode = {
          type: 'REPEAT',
          repeat: {
            repeat_effect: node.atomic_effect
              ? {
                  type: 'THEN',
                  atomic_effect: node.atomic_effect,
                  sequential: { next_id: null, next: null },
                }
              : createDefaultEffectNode('THEN'),
            count: 1,
          },
        }
        break
      case 'FOREACH':
        newNode = {
          type: 'FOREACH',
          for_each: {
            for_each_effect: node.atomic_effect
              ? {
                  type: 'THEN',
                  atomic_effect: node.atomic_effect,
                  sequential: { next_id: null, next: null },
                }
              : createDefaultEffectNode('THEN'),
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
        break
      default:
        newNode = node
    }
    onChange(newNode)
  }

  return (
    <div
      className="effect-node-editor"
      style={{ marginLeft: `${depth * 20}px` }}
    >
      <div className="effect-node-header">
        <select
          value={node.type}
          onChange={(e) => updateNodeType(e.target.value as EffectNodeType)}
        >
          <option value="THEN">順次実行 (THEN)</option>
          <option value="AND">並列実行 (AND)</option>
          <option value="IF_ELSE">条件分岐 (IF_ELSE)</option>
          <option value="REPEAT">繰り返し (REPEAT)</option>
          <option value="FOREACH">反復 (FOREACH)</option>
        </select>
        {onDelete && (
          <button type="button" onClick={onDelete}>
            削除
          </button>
        )}
      </div>

      {/* THEN ノード */}
      {node.type === 'THEN' && (
        <div className="effect-node-content">
          {node.atomic_effect && (
            <AtomicEffectEditor
              effect={node.atomic_effect}
              onChange={(newEffect) => {
                onChange({
                  ...node,
                  atomic_effect: newEffect,
                })
              }}
            />
          )}
          {node.sequential && (
            <div className="effect-node-next">
              <div className="form-label">次の効果:</div>
              {node.sequential.next ? (
                <EffectNodeEditor
                  node={node.sequential.next}
                  onChange={(newNext) => {
                    if (node.sequential) {
                      onChange({
                        ...node,
                        sequential: {
                          ...node.sequential,
                          next: newNext,
                        },
                      })
                    }
                  }}
                  onDelete={() => {
                    if (node.sequential) {
                      onChange({
                        ...node,
                        sequential: {
                          ...node.sequential,
                          next: null,
                        },
                      })
                    }
                  }}
                  depth={depth + 1}
                />
              ) : (
                <button
                  type="button"
                  onClick={() => {
                    if (node.sequential) {
                      onChange({
                        ...node,
                        sequential: {
                          ...node.sequential,
                          next: {
                            type: 'THEN',
                            atomic_effect: createDefaultAtomicEffect(),
                            sequential: { next_id: null, next: null },
                          },
                        },
                      })
                    }
                  }}
                >
                  + 次の効果を追加
                </button>
              )}
            </div>
          )}
        </div>
      )}

      {/* AND ノード */}
      {node.type === 'AND' && node.parallel && (
        <div className="effect-node-content">
          <div className="effect-node-children">
            <div className="form-label">並列実行する効果:</div>
            {node.parallel.children.map((child, index) => (
              <EffectNodeEditor
                // biome-ignore lint/suspicious/noArrayIndexKey: 並列ノードの子要素は順序が変わらないため
                key={`parallel-${depth}-${index}`}
                node={child}
                onChange={(newChild) => {
                  if (node.parallel) {
                    const newChildren = [...node.parallel.children]
                    newChildren[index] = newChild
                    onChange({
                      ...node,
                      parallel: {
                        ...node.parallel,
                        children: newChildren,
                      },
                    })
                  }
                }}
                onDelete={() => {
                  if (node.parallel) {
                    const newChildren = node.parallel.children.filter(
                      (_, i) => i !== index,
                    )
                    onChange({
                      ...node,
                      parallel: {
                        ...node.parallel,
                        children: newChildren,
                      },
                    })
                  }
                }}
                depth={depth + 1}
              />
            ))}
            <button
              type="button"
              onClick={() => {
                if (node.parallel) {
                  onChange({
                    ...node,
                    parallel: {
                      ...node.parallel,
                      children: [
                        ...node.parallel.children,
                        {
                          type: 'THEN',
                          atomic_effect: createDefaultAtomicEffect(),
                          sequential: { next_id: null, next: null },
                        },
                      ],
                    },
                  })
                }
              }}
            >
              + 効果を追加
            </button>
          </div>
        </div>
      )}

      {/* IF_ELSE ノード */}
      {node.type === 'IF_ELSE' && node.if_else && (
        <div className="effect-node-content">
          <div className="effect-node-condition">
            <div className="form-label">条件:</div>
            <div className="condition-editor">
              <select
                value={node.if_else.condition.type || 'PLAYER_HP'}
                onChange={(e) => {
                  if (node.if_else) {
                    onChange({
                      ...node,
                      if_else: {
                        ...node.if_else,
                        condition: {
                          ...node.if_else.condition,
                          type: e.target.value,
                        },
                      },
                    })
                  }
                }}
              >
                <option value="PLAYER_HP">プレイヤーHP</option>
                <option value="PLAYER_MANA">プレイヤーマナ</option>
                <option value="UNIT_COUNT">ユニット数</option>
                <option value="HAND_SIZE">手札サイズ</option>
                <option value="DECK_SIZE">デッキサイズ</option>
                <option value="TURN_NUMBER">ターン数</option>
                <option value="UNIT_ATTACK">ユニット攻撃力</option>
                <option value="UNIT_DEFENSE">ユニット防御力</option>
                <option value="HAS_KEYWORD">特性を持っている</option>
                <option value="CARD_PLAYED">カードを使用した</option>
                <option value="DAMAGE_TAKEN">ダメージを受けた</option>
              </select>
              <select
                value={node.if_else.condition.operator || 'GREATER_THAN'}
                onChange={(e) => {
                  if (node.if_else) {
                    onChange({
                      ...node,
                      if_else: {
                        ...node.if_else,
                        condition: {
                          ...node.if_else.condition,
                          operator: e.target.value,
                        },
                      },
                    })
                  }
                }}
              >
                <option value="EQUAL">等しい</option>
                <option value="NOT_EQUAL">等しくない</option>
                <option value="LESS_THAN">より小さい</option>
                <option value="GREATER_THAN">より大きい</option>
                <option value="LESS_THAN_OR_EQUAL">以下</option>
                <option value="GREATER_THAN_OR_EQUAL">以上</option>
              </select>
              <input
                type="number"
                value={node.if_else.condition.value || 0}
                onChange={(e) => {
                  if (node.if_else) {
                    onChange({
                      ...node,
                      if_else: {
                        ...node.if_else,
                        condition: {
                          ...node.if_else.condition,
                          value: Number(e.target.value),
                        },
                      },
                    })
                  }
                }}
              />
            </div>
          </div>
          <div className="effect-node-then">
            <div className="form-label">Then (条件が真の場合):</div>
            <EffectNodeEditor
              node={node.if_else.thenNode}
              onChange={(newThen) => {
                if (node.if_else) {
                  onChange({
                    ...node,
                    if_else: {
                      ...node.if_else,
                      thenNode: newThen,
                    },
                  })
                }
              }}
              depth={depth + 1}
            />
          </div>
          <div className="effect-node-else">
            <div className="form-label">Else (条件が偽の場合):</div>
            {node.if_else.elseNode ? (
              <EffectNodeEditor
                node={node.if_else.elseNode}
                onChange={(newElse) => {
                  if (node.if_else) {
                    onChange({
                      ...node,
                      if_else: {
                        ...node.if_else,
                        elseNode: newElse,
                      },
                    })
                  }
                }}
                onDelete={() => {
                  if (node.if_else) {
                    onChange({
                      ...node,
                      if_else: {
                        ...node.if_else,
                        elseNode: null,
                      },
                    })
                  }
                }}
                depth={depth + 1}
              />
            ) : (
              <button
                type="button"
                onClick={() => {
                  if (node.if_else) {
                    onChange({
                      ...node,
                      if_else: {
                        ...node.if_else,
                        elseNode: {
                          type: 'THEN',
                          atomic_effect: createDefaultAtomicEffect(),
                          sequential: { next_id: null, next: null },
                        },
                      },
                    })
                  }
                }}
              >
                + Elseを追加
              </button>
            )}
          </div>
        </div>
      )}

      {/* REPEAT ノード */}
      {node.type === 'REPEAT' && node.repeat && (
        <div className="effect-node-content">
          <div className="form-group">
            <label htmlFor={`repeat-count-${depth}`}>繰り返し回数:</label>
            <input
              id={`repeat-count-${depth}`}
              type="number"
              min="1"
              value={node.repeat.count}
              onChange={(e) => {
                if (node.repeat) {
                  onChange({
                    ...node,
                    repeat: {
                      ...node.repeat,
                      count: Number(e.target.value),
                    },
                  })
                }
              }}
            />
          </div>
          <div className="effect-node-repeat-effect">
            <div className="form-label">繰り返す効果:</div>
            <EffectNodeEditor
              node={node.repeat.repeat_effect}
              onChange={(newEffect) => {
                if (node.repeat) {
                  onChange({
                    ...node,
                    repeat: {
                      ...node.repeat,
                      repeat_effect: newEffect,
                    },
                  })
                }
              }}
              depth={depth + 1}
            />
          </div>
        </div>
      )}

      {/* FOREACH ノード */}
      {node.type === 'FOREACH' && node.for_each && (
        <div className="effect-node-content">
          <div className="effect-node-foreach-target">
            <label htmlFor={`foreach-target-${depth}`}>対象:</label>
            <select
              id={`foreach-target-${depth}`}
              value={node.for_each.for_each_target.type}
              onChange={(e) => {
                if (node.for_each) {
                  onChange({
                    ...node,
                    for_each: {
                      ...node.for_each,
                      for_each_target: {
                        ...node.for_each.for_each_target,
                        type: e.target.value,
                      },
                    },
                  })
                }
              }}
            >
              <option value="Self">自分自身</option>
              <option value="Opponent">相手プレイヤー</option>
              <option value="Allies">自分側の全ユニット</option>
              <option value="Enemies">相手側の全ユニット</option>
              <option value="AllUnits">両者の全ユニット</option>
              <option value="RandomAlly">ランダムな味方ユニット</option>
              <option value="RandomEnemy">ランダムな敵ユニット</option>
              <option value="Specific">明示的に指定されたユニット</option>
            </select>
          </div>
          <div className="effect-node-foreach-effect">
            <div className="form-label">各対象に適用する効果:</div>
            <EffectNodeEditor
              node={node.for_each.for_each_effect}
              onChange={(newEffect) => {
                if (node.for_each) {
                  onChange({
                    ...node,
                    for_each: {
                      ...node.for_each,
                      for_each_effect: newEffect,
                    },
                  })
                }
              }}
              depth={depth + 1}
            />
          </div>
        </div>
      )}
    </div>
  )
}

// AtomicEffectEditor コンポーネント
interface AtomicEffectEditorProps {
  effect: AtomicEffect
  onChange: (effect: AtomicEffect) => void
}

function AtomicEffectEditor({ effect, onChange }: AtomicEffectEditorProps) {
  return (
    <div className="atomic-effect-editor">
      <div className="form-group">
        <label htmlFor="atomic-effect-type">効果タイプ:</label>
        <select
          id="atomic-effect-type"
          value={effect.type}
          onChange={(e) => onChange({ ...effect, type: e.target.value })}
        >
          <option value="DEAL_DAMAGE">ダメージを与える</option>
          <option value="DEAL_SPLASH">範囲ダメージ</option>
          <option value="RESTORE_HP">HP回復</option>
          <option value="RESTORE_MANA">マナ回復</option>
          <option value="FULL_RESTORE">完全回復</option>
          <option value="DRAW_CARD">カードドロー</option>
          <option value="DISCARD_CARD">手札を捨てる</option>
          <option value="SEARCH_CARD">デッキからサーチ</option>
          <option value="SHUFFLE_DECK">デッキシャッフル</option>
          <option value="MODIFY_ATTACK">攻撃力変更</option>
          <option value="MODIFY_DEFENSE">防御力変更</option>
          <option value="MODIFY_COST">コスト変更</option>
          <option value="MODIFY_MAX_HP">最大HP変更</option>
          <option value="SUMMON_UNIT">ユニット召喚</option>
          <option value="DESTROY_UNIT">ユニット破壊</option>
          <option value="RETURN_TO_HAND">手札に戻す</option>
          <option value="RETURN_TO_DECK">デッキに戻す</option>
          <option value="DISABLE_UNIT">効果無効化</option>
          <option value="GRANT_TRAIT">特性付与</option>
          <option value="REMOVE_TRAIT">特性除去</option>
          <option value="GAIN_MANA">マナ増加</option>
          <option value="REDUCE_COST">コスト減少</option>
        </select>
      </div>
      <div className="form-group">
        <label htmlFor="atomic-effect-value">効果値:</label>
        <input
          id="atomic-effect-value"
          type="number"
          value={effect.value}
          onChange={(e) =>
            onChange({ ...effect, value: Number(e.target.value) })
          }
        />
      </div>
      <div className="form-group">
        <label htmlFor="atomic-effect-target">対象:</label>
        <select
          id="atomic-effect-target"
          value={effect.target.type}
          onChange={(e) =>
            onChange({
              ...effect,
              target: { ...effect.target, type: e.target.value },
            })
          }
        >
          <option value="Self">自分自身</option>
          <option value="Opponent">相手プレイヤー</option>
          <option value="Allies">自分側の全ユニット</option>
          <option value="Enemies">相手側の全ユニット</option>
          <option value="AllUnits">両者の全ユニット</option>
          <option value="RandomAlly">ランダムな味方ユニット</option>
          <option value="RandomEnemy">ランダムな敵ユニット</option>
          <option value="Specific">明示的に指定されたユニット</option>
        </select>
      </div>
      <div className="form-group">
        <label htmlFor="atomic-effect-timing">発動タイミング:</label>
        <select
          id="atomic-effect-timing"
          value={effect.timing}
          onChange={(e) => onChange({ ...effect, timing: e.target.value })}
        >
          <option value="Immediate">即時発動</option>
          <option value="OnSummon">召喚時に発動</option>
          <option value="OnDestroy">破壊時に発動</option>
          <option value="OnAttack">攻撃時に発動</option>
          <option value="OnDamaged">ダメージを受けたときに発動</option>
          <option value="TurnStart">ターン開始時</option>
          <option value="TurnEnd">ターン終了時</option>
        </select>
      </div>
    </div>
  )
}

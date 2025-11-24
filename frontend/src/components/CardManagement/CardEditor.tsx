import { useEffect, useState } from 'react'
import { CardManagementService } from '../../gen/card_management_connect'
import type { Card, Trait } from '../../gen/common_pb'
import { CardType } from '../../gen/common_pb'
import { createAuthenticatedClient } from '../../lib/auth'
import './CardEditor.css'

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
    cardEffectJson: '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const cardClient = createAuthenticatedClient(CardManagementService)

  useEffect(() => {
    if (card) {
      setFormData({
        id: card.id,
        name: card.name,
        type: card.type,
        cost: card.cost,
        attack: card.attack,
        defense: card.defense,
        effectText: card.effect,
        traits: card.traits || [],
        cardEffectJson: '', // CardEffectはJSONとして管理
      })
    } else if (isNewCardMode) {
      // 新しいカード作成モード
      setFormData({
        id: '',
        name: '',
        type: CardType.UNIT,
        cost: 0,
        attack: undefined,
        defense: undefined,
        effectText: '',
        traits: [],
        cardEffectJson: '',
      })
    }
  }, [card, isNewCardMode])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      if (card) {
        // 更新
        await cardClient.updateCard({
          id: formData.id,
          name: formData.name,
          type: formData.type,
          cost: formData.cost,
          attack: formData.attack,
          defense: formData.defense,
          effectText: formData.effectText,
          traits: formData.traits,
          cardEffectJson: formData.cardEffectJson,
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
          effectText: formData.effectText,
          traits: formData.traits,
          cardEffectJson: formData.cardEffectJson,
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
            onChange={(e) =>
              setFormData({
                ...formData,
                type: Number(e.target.value) as CardType,
              })
            }
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
        <div className="form-group">
          <label htmlFor="effectText">効果テキスト</label>
          <textarea
            id="effectText"
            value={formData.effectText}
            onChange={(e) =>
              setFormData({ ...formData, effectText: e.target.value })
            }
            rows={4}
          />
        </div>
        <div className="form-group">
          <label htmlFor="cardEffectJson">カード効果JSON（上級者向け）</label>
          <textarea
            id="cardEffectJson"
            value={formData.cardEffectJson}
            onChange={(e) =>
              setFormData({ ...formData, cardEffectJson: e.target.value })
            }
            rows={8}
            placeholder='{"definitions":[...]}'
          />
        </div>
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

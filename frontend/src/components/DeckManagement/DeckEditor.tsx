import { useEffect, useState } from 'react'
import type { Card, Deck } from '../../gen/common_pb'
import { cardManagementClient } from '../../lib/api-client'
import './DeckEditor.css'

const DECK_SIZE = 40
const MAX_COPIES_PER_CARD = 3

interface DeckEditorProps {
  deck: Deck | null
  isNewDeckMode: boolean
  onSave: (deckId: string) => void
  onCancel: () => void
}

interface DeckCardEntry {
  id: string // ユニークID
  cardId: string // カードID
}

export default function DeckEditor({
  deck,
  isNewDeckMode,
  onSave,
  onCancel,
}: DeckEditorProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [deckCards, setDeckCards] = useState<DeckCardEntry[]>([])
  const [availableCards, setAvailableCards] = useState<Card[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // カード一覧の読み込み
  useEffect(() => {
    const loadCards = async () => {
      try {
        setLoading(true)
        const response = await cardManagementClient.listCards({})
        const cards = response.cards || []
        console.log('カードを読み込みました:', cards)
        // カードデータの詳細をログ出力
        if (cards.length > 0) {
          console.log('サンプルカードデータ:', {
            id: cards[0].id,
            name: cards[0].name,
            cost: cards[0].cost,
            type: cards[0].type,
            attack: cards[0].attack,
            defense: cards[0].defense,
            effect: cards[0].effect,
            traits: cards[0].traits,
            cardEffect: cards[0].cardEffect,
          })
        }
        setAvailableCards(cards)
      } catch (err: unknown) {
        console.error('カードの読み込みに失敗:', err)
      } finally {
        setLoading(false)
      }
    }
    loadCards()
  }, [])

  // デッキ情報の読み込み
  useEffect(() => {
    if (deck) {
      setName(deck.name)
      setDescription(deck.description)
      setDeckCards(
        deck.cardIds.map((cardId, idx) => ({
          id: `${cardId}-${idx}-${Date.now()}`,
          cardId,
        })),
      )
    } else {
      setName('')
      setDescription('')
      setDeckCards([])
    }
    setError(null)
  }, [deck])

  const handleSave = async () => {
    if (!name.trim()) {
      setError('デッキ名を入力してください')
      return
    }

    if (deckCards.length !== DECK_SIZE) {
      setError(
        `デッキは${DECK_SIZE}枚である必要があります（現在${deckCards.length}枚）`,
      )
      return
    }

    try {
      setSaving(true)
      setError(null)

      let savedDeckId: string

      if (isNewDeckMode) {
        const response = await cardManagementClient.createDeck({
          name: name.trim(),
          description: description.trim(),
          cardIds: deckCards.map((entry) => entry.cardId),
        })
        savedDeckId = response.deck?.id || ''
      } else if (deck) {
        const response = await cardManagementClient.updateDeck({
          id: deck.id,
          name: name.trim(),
          description: description.trim(),
          cardIds: deckCards.map((entry) => entry.cardId),
        })
        savedDeckId = response.deck?.id || deck.id
      } else {
        return
      }

      onSave(savedDeckId)
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'デッキの保存に失敗しました',
      )
    } finally {
      setSaving(false)
    }
  }

  const handleAddCard = (cardId: string) => {
    // 40枚制限チェック
    if (deckCards.length >= DECK_SIZE) {
      setError(`デッキは${DECK_SIZE}枚までです`)
      return
    }

    // 同じカード3枚制限チェック
    const currentCount = getCardCount(cardId)
    if (currentCount >= MAX_COPIES_PER_CARD) {
      setError(`同じカードは${MAX_COPIES_PER_CARD}枚までです`)
      return
    }

    const newEntry: DeckCardEntry = {
      id: `${cardId}-${Date.now()}-${Math.random()}`,
      cardId,
    }
    setDeckCards((prev) => [...prev, newEntry])
    setError(null)
  }

  const handleRemoveCard = (entryId: string) => {
    setDeckCards((prev) => prev.filter((entry) => entry.id !== entryId))
    setError(null)
  }

  const getCardById = (cardId: string): Card | undefined => {
    return availableCards.find((c) => c.id === cardId)
  }

  const getCardCount = (cardId: string): number => {
    return deckCards.filter((entry) => entry.cardId === cardId).length
  }

  if (!isNewDeckMode && !deck) {
    return (
      <div className="deck-editor-empty">
        デッキを選択するか、新しいデッキを作成してください
      </div>
    )
  }

  return (
    <div className="deck-editor">
      <h2>{isNewDeckMode ? '新しいデッキ' : 'デッキ編集'}</h2>

      {error && <div className="deck-editor-error">{error}</div>}

      <div className="deck-editor-form">
        <div className="form-group">
          <label htmlFor="deck-name">
            デッキ名 <span className="required">*</span>
          </label>
          <input
            id="deck-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="デッキ名を入力"
            maxLength={50}
          />
        </div>

        <div className="form-group">
          <label htmlFor="deck-description">説明</label>
          <textarea
            id="deck-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="デッキの説明を入力（任意）"
            rows={3}
            maxLength={200}
          />
        </div>

        <div className="deck-cards-section">
          <div className="deck-current-cards">
            <div className="deck-header">
              <h3>
                デッキ内のカード ({deckCards.length}/{DECK_SIZE}枚)
              </h3>
              {deckCards.length < DECK_SIZE && (
                <span className="deck-remaining-count warning">
                  あと{DECK_SIZE - deckCards.length}枚必要
                </span>
              )}
              {deckCards.length === DECK_SIZE && (
                <span className="deck-remaining-count complete">✓ 完成</span>
              )}
            </div>
            {deckCards.length === 0 ? (
              <div className="empty-message">カードが追加されていません</div>
            ) : (
              <div className="deck-card-list">
                {deckCards.map((entry) => {
                  const card = getCardById(entry.cardId)
                  if (!card) return null
                  const effectText = card.effect || ''
                  const traitsText = card.traits?.join(', ') || ''

                  return (
                    <div
                      key={entry.id}
                      className="deck-card-item"
                      title={`${card.name}${effectText ? `\n効果: ${effectText}` : ''}${traitsText ? `\n特性: ${traitsText}` : ''}`}
                    >
                      <div className="deck-card-main">
                        <span className="deck-card-name">{card.name}</span>
                        <div className="deck-card-details">
                          <span className="deck-card-cost">
                            コスト: {card.cost}
                          </span>
                          {card.attack !== undefined &&
                            card.defense !== undefined && (
                              <span className="deck-card-stats">
                                {card.attack}/{card.defense}
                              </span>
                            )}
                        </div>
                        {effectText && (
                          <div className="deck-card-effect">{effectText}</div>
                        )}
                      </div>
                      <button
                        type="button"
                        className="deck-card-remove"
                        onClick={() => handleRemoveCard(entry.id)}
                      >
                        ×
                      </button>
                    </div>
                  )
                })}
              </div>
            )}
          </div>

          <div className="deck-available-cards">
            <h3>利用可能なカード</h3>
            {loading ? (
              <div className="loading-message">読み込み中...</div>
            ) : availableCards.length === 0 ? (
              <div className="empty-message">カードがありません</div>
            ) : (
              <div className="available-card-list">
                {availableCards.map((card) => {
                  const count = getCardCount(card.id)
                  const isMaxCopies = count >= MAX_COPIES_PER_CARD
                  const isDeckFull = deckCards.length >= DECK_SIZE
                  const canAdd = !isMaxCopies && !isDeckFull
                  const effectText = card.effect || ''
                  const traitsText = card.traits?.join(', ') || ''

                  return (
                    <div
                      key={card.id}
                      className={`available-card-item ${!canAdd ? 'disabled' : ''}`}
                      title={`${card.name}${effectText ? `\n効果: ${effectText}` : ''}${traitsText ? `\n特性: ${traitsText}` : ''}`}
                    >
                      <div className="available-card-info">
                        <div className="available-card-header">
                          <span className="available-card-name">
                            {card.name}
                          </span>
                          <div className="available-card-meta">
                            <span className="available-card-cost">
                              コスト: {card.cost}
                            </span>
                            {card.attack !== undefined &&
                              card.defense !== undefined && (
                                <span className="available-card-stats">
                                  {card.attack}/{card.defense}
                                </span>
                              )}
                            {count > 0 && (
                              <span
                                className={`available-card-count ${isMaxCopies ? 'max' : ''}`}
                              >
                                ×{count}/{MAX_COPIES_PER_CARD}
                              </span>
                            )}
                          </div>
                        </div>
                        {effectText && (
                          <div className="available-card-effect">
                            {effectText}
                          </div>
                        )}
                      </div>
                      <button
                        type="button"
                        className="available-card-add"
                        onClick={() => handleAddCard(card.id)}
                        disabled={!canAdd}
                        title={
                          isDeckFull
                            ? `デッキは${DECK_SIZE}枚までです`
                            : isMaxCopies
                              ? `同じカードは${MAX_COPIES_PER_CARD}枚までです`
                              : '追加'
                        }
                      >
                        {canAdd ? '追加' : '上限'}
                      </button>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </div>

        <div className="deck-editor-actions">
          <button
            type="button"
            className="btn-cancel"
            onClick={onCancel}
            disabled={saving}
          >
            キャンセル
          </button>
          <button
            type="button"
            className="btn-save"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>
    </div>
  )
}

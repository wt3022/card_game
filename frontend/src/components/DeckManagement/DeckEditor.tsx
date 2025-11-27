import { useEffect, useState } from 'react'
import { TRAIT_LABELS } from '../../constants/game'
import type { Card, Deck } from '../../gen/common_pb'
import { CardType } from '../../gen/common_pb'
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
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<'name' | 'cost' | 'attack' | 'defense'>(
    'name',
  )
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')
  const [deckSortBy, setDeckSortBy] = useState<
    'name' | 'cost' | 'attack' | 'defense'
  >('cost')
  const [deckSortOrder, setDeckSortOrder] = useState<'asc' | 'desc'>('asc')
  const [typeFilter, setTypeFilter] = useState<CardType | 'ALL'>('ALL')
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const pageSize = 50

  // カード一覧の読み込み
  useEffect(() => {
    const loadCards = async () => {
      try {
        setLoading(true)
        const response = await cardManagementClient.listCards({
          page: currentPage,
          pageSize: pageSize,
        })
        const cards = response.cards || []
        setTotalPages(response.totalPages)
        setTotalCount(response.totalCount)
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
  }, [currentPage])

  // デッキ情報の読み込み
  useEffect(() => {
    if (deck) {
      console.log('デッキ情報を読み込みました:', {
        id: deck.id,
        name: deck.name,
        cardIds: deck.cardIds,
        cardIdsLength: deck.cardIds.length,
      })
      console.log('最初の5枚のカードID:', deck.cardIds.slice(0, 5))
      console.log('最後の5枚のカードID:', deck.cardIds.slice(-5))
      setName(deck.name)
      setDescription(deck.description)
      const cards = deck.cardIds.map((cardId, idx) => ({
        id: `${cardId}-${idx}-${Date.now()}`,
        cardId,
      }))
      console.log('デッキカードエントリー:', cards.length, '枚')
      console.log('最初の3エントリー:', cards.slice(0, 3))
      console.log('最後の3エントリー:', cards.slice(-3))
      setDeckCards(cards)
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
    const found = availableCards.find((c) => c.id === cardId)
    if (!found && availableCards.length > 0) {
      console.warn('カードが見つかりません:', {
        searchCardId: cardId,
        availableCardIds: availableCards.slice(0, 5).map((c) => c.id),
        totalAvailable: availableCards.length,
      })
    }
    return found
  }

  const getCardCount = (cardId: string): number => {
    return deckCards.filter((entry) => entry.cardId === cardId).length
  }

  // カードのフィルタリング
  const filterCards = (cards: Card[]): Card[] => {
    let filtered = cards

    // タイプフィルター
    if (typeFilter !== 'ALL') {
      filtered = filtered.filter((card) => card.type === typeFilter)
    }

    // テキスト検索
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter((card) => {
        // カード名での検索
        if (card.name.toLowerCase().includes(query)) return true

        // 効果テキストでの検索
        if (card.effect?.toLowerCase().includes(query)) return true

        // 特性での検索
        if (
          card.traits?.some((trait) =>
            TRAIT_LABELS[trait]?.toLowerCase().includes(query),
          )
        )
          return true

        return false
      })
    }

    return filtered
  }

  // カードのソート
  const sortCards = (cards: Card[]): Card[] => {
    const sorted = [...cards]
    sorted.sort((a, b) => {
      let compareValue = 0

      switch (sortBy) {
        case 'name':
          compareValue = a.name.localeCompare(b.name, 'ja')
          break
        case 'cost':
          compareValue = (a.cost || 0) - (b.cost || 0)
          break
        case 'attack':
          compareValue = (a.attack || 0) - (b.attack || 0)
          break
        case 'defense':
          compareValue = (a.defense || 0) - (b.defense || 0)
          break
      }

      return sortOrder === 'asc' ? compareValue : -compareValue
    })

    return sorted
  }

  // フィルタリングとソートを適用したカード一覧
  const filteredAndSortedCards = sortCards(filterCards(availableCards))

  // デッキ内のカードをグループ化（同じカードをまとめる）
  const groupedDeckCards = deckCards.reduce(
    (acc, entry) => {
      const existing = acc.find((group) => group.cardId === entry.cardId)
      if (existing) {
        existing.count++
        existing.entries.push(entry)
      } else {
        acc.push({
          cardId: entry.cardId,
          count: 1,
          entries: [entry],
        })
      }
      return acc
    },
    [] as Array<{ cardId: string; count: number; entries: DeckCardEntry[] }>,
  )

  // デッキ内カードをソート
  const sortedGroupedDeckCards = [...groupedDeckCards].sort((a, b) => {
    const cardA = getCardById(a.cardId)
    const cardB = getCardById(b.cardId)
    if (!cardA || !cardB) return 0

    let compareValue = 0
    switch (deckSortBy) {
      case 'name':
        compareValue = cardA.name.localeCompare(cardB.name, 'ja')
        break
      case 'cost':
        compareValue = (cardA.cost || 0) - (cardB.cost || 0)
        break
      case 'attack':
        compareValue = (cardA.attack || 0) - (cardB.attack || 0)
        break
      case 'defense':
        compareValue = (cardB.defense || 0) - (cardB.defense || 0)
        break
    }
    return deckSortOrder === 'asc' ? compareValue : -compareValue
  })

  // マナカーブの統計
  const manaCurve = deckCards.reduce(
    (acc, entry) => {
      const card = getCardById(entry.cardId)
      if (card) {
        const cost = Math.min(card.cost || 0, 10) // 10以上は10にまとめる
        acc[cost] = (acc[cost] || 0) + 1
      }
      return acc
    },
    {} as Record<number, number>,
  )

  // カード種別の統計
  const typeDistribution = deckCards.reduce(
    (acc, entry) => {
      const card = getCardById(entry.cardId)
      if (card) {
        const type = card.type
        const typeName =
          type === CardType.UNIT
            ? 'ユニット'
            : type === CardType.SPELL
              ? 'スペル'
              : type === CardType.LEADER
                ? 'リーダー'
                : '不明'
        acc[typeName] = (acc[typeName] || 0) + 1
      }
      return acc
    },
    {} as Record<string, number>,
  )

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

        {/* デッキ分析 */}
        <div className="deck-stats-container">
          <div className="deck-stats">
            <div className="deck-stats-section">
              <h4>マナカーブ</h4>
              <div className="mana-curve">
                {Array.from({ length: 11 }, (_, i) => {
                  const count = manaCurve[i] || 0
                  const maxCount = Math.max(...Object.values(manaCurve), 1)
                  return (
                    <div key={`mana-${i}`} className="mana-bar-container">
                      <div className="mana-count">{count}</div>
                      <div className="mana-bar-wrapper">
                        <div
                          className="mana-bar"
                          style={{
                            height: `${(count / maxCount) * 100}%`,
                          }}
                        />
                      </div>
                      <div className="mana-cost-label">{i}</div>
                    </div>
                  )
                })}
              </div>
            </div>

            <div className="deck-stats-section">
              <h4>カード種別</h4>
              {deckCards.length > 0 ? (
                <div className="type-distribution">
                  {Object.entries(typeDistribution)
                    .sort(([, a], [, b]) => b - a)
                    .map(([type, count]) => (
                      <div key={type} className="type-item">
                        <span className="type-label">{type}</span>
                        <span className="type-count">{count}枚</span>
                        <div className="type-bar">
                          <div
                            className="type-bar-fill"
                            style={{
                              width: `${(count / deckCards.length) * 100}%`,
                            }}
                          />
                        </div>
                      </div>
                    ))}
                </div>
              ) : (
                <div className="empty-stats">
                  カードを追加すると統計が表示されます
                </div>
              )}
            </div>
          </div>
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

            {/* デッキ内カードのソート */}
            {deckCards.length > 0 && (
              <div className="deck-sort-controls">
                <select
                  value={deckSortBy}
                  onChange={(e) =>
                    setDeckSortBy(e.target.value as typeof deckSortBy)
                  }
                  className="sort-select"
                >
                  <option value="cost">コスト順</option>
                  <option value="name">名前順</option>
                  <option value="attack">攻撃力順</option>
                  <option value="defense">防御力順</option>
                </select>
                <button
                  type="button"
                  className="sort-order-btn"
                  onClick={() =>
                    setDeckSortOrder(deckSortOrder === 'asc' ? 'desc' : 'asc')
                  }
                  title={deckSortOrder === 'asc' ? '昇順' : '降順'}
                >
                  {deckSortOrder === 'asc' ? '↑' : '↓'}
                </button>
              </div>
            )}
            {deckCards.length === 0 ? (
              <div className="empty-message">カードが追加されていません</div>
            ) : (
              <div className="deck-card-list">
                {sortedGroupedDeckCards.map((group) => {
                  const card = getCardById(group.cardId)
                  if (!card) {
                    console.warn(`Card not found:`, group.cardId)
                    return null
                  }
                  const effectText = card.effect || ''

                  return (
                    <div key={group.cardId} className="deck-card-item">
                      <div className="deck-card-main">
                        <div className="deck-card-header">
                          <span className="deck-card-name">{card.name}</span>
                          <div className="deck-card-header-right">
                            <span className="deck-card-count">
                              ×{group.count}
                            </span>
                            <span className="deck-card-cost">{card.cost}</span>
                          </div>
                        </div>
                        <div className="deck-card-details">
                          {card.attack !== undefined &&
                            card.defense !== undefined && (
                              <span className="deck-card-stats">
                                ATK {card.attack} / DEF {card.defense}
                              </span>
                            )}
                        </div>
                        {card.traits && card.traits.length > 0 && (
                          <div className="deck-card-traits">
                            {card.traits.map((trait) => (
                              <span key={trait} className="trait-badge">
                                {TRAIT_LABELS[trait] || trait}
                              </span>
                            ))}
                          </div>
                        )}
                        {effectText && (
                          <div className="deck-card-effect">{effectText}</div>
                        )}
                      </div>
                      <div className="deck-card-actions">
                        <button
                          type="button"
                          className="deck-card-add"
                          onClick={() => handleAddCard(group.cardId)}
                          disabled={
                            group.count >= MAX_COPIES_PER_CARD ||
                            deckCards.length >= DECK_SIZE
                          }
                          title="1枚追加"
                        >
                          +
                        </button>
                        <button
                          type="button"
                          className="deck-card-remove"
                          onClick={() => handleRemoveCard(group.entries[0].id)}
                          title="1枚削除"
                        >
                          −
                        </button>
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>

          <div className="deck-available-cards">
            <h3>利用可能なカード</h3>

            {/* タイプフィルター */}
            <div className="type-filter-buttons">
              <button
                type="button"
                className={`type-filter-btn ${typeFilter === 'ALL' ? 'active' : ''}`}
                onClick={() => setTypeFilter('ALL')}
              >
                全て
              </button>
              <button
                type="button"
                className={`type-filter-btn ${typeFilter === CardType.UNIT ? 'active' : ''}`}
                onClick={() => setTypeFilter(CardType.UNIT)}
              >
                ユニット
              </button>
              <button
                type="button"
                className={`type-filter-btn ${typeFilter === CardType.SPELL ? 'active' : ''}`}
                onClick={() => setTypeFilter(CardType.SPELL)}
              >
                スペル
              </button>
            </div>

            {/* 検索とソートのコントロール */}
            <div className="card-controls">
              <div className="search-box">
                <input
                  type="text"
                  placeholder="カード名、効果、特性で検索..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="search-input"
                />
                {searchQuery && (
                  <button
                    type="button"
                    className="search-clear"
                    onClick={() => setSearchQuery('')}
                    title="クリア"
                  >
                    ×
                  </button>
                )}
              </div>

              <div className="sort-controls">
                <select
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value as typeof sortBy)}
                  className="sort-select"
                >
                  <option value="name">名前</option>
                  <option value="cost">コスト</option>
                  <option value="attack">攻撃力</option>
                  <option value="defense">防御力</option>
                </select>
                <button
                  type="button"
                  className="sort-order-btn"
                  onClick={() =>
                    setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
                  }
                  title={sortOrder === 'asc' ? '昇順' : '降順'}
                >
                  {sortOrder === 'asc' ? '↑' : '↓'}
                </button>
              </div>
            </div>

            {loading ? (
              <div className="loading-message">読み込み中...</div>
            ) : availableCards.length === 0 ? (
              <div className="empty-message">カードがありません</div>
            ) : filteredAndSortedCards.length === 0 ? (
              <div className="empty-message">
                検索条件に一致するカードがありません
              </div>
            ) : (
              <>
                <div className="available-card-list">
                  {filteredAndSortedCards.map((card) => {
                    const count = getCardCount(card.id)
                    const isMaxCopies = count >= MAX_COPIES_PER_CARD
                    const isDeckFull = deckCards.length >= DECK_SIZE
                    const canAdd = !isMaxCopies && !isDeckFull
                    const effectText = card.effect || ''

                    return (
                      <div
                        key={card.id}
                        className={`available-card-item ${!canAdd ? 'disabled' : ''}`}
                      >
                        <div className="available-card-info">
                          <div className="available-card-header">
                            <span className="available-card-name">
                              {card.name}
                            </span>
                            <span className="available-card-cost">
                              {card.cost}
                            </span>
                          </div>
                          <div className="available-card-meta">
                            {card.attack !== undefined &&
                              card.defense !== undefined && (
                                <span className="available-card-stats">
                                  ATK {card.attack} / DEF {card.defense}
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
                          {card.traits && card.traits.length > 0 && (
                            <div className="available-card-traits">
                              {card.traits.map((trait) => (
                                <span key={trait} className="trait-badge">
                                  {TRAIT_LABELS[trait] || trait}
                                </span>
                              ))}
                            </div>
                          )}
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
                <div className="pagination">
                  <button
                    type="button"
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1}
                    className="pagination-btn"
                  >
                    ← 前へ
                  </button>
                  <span className="pagination-info">
                    {currentPage} / {totalPages} ページ (全{totalCount}件)
                  </span>
                  <button
                    type="button"
                    onClick={() =>
                      setCurrentPage((p) => Math.min(totalPages, p + 1))
                    }
                    disabled={currentPage >= totalPages}
                    className="pagination-btn"
                  >
                    次へ →
                  </button>
                </div>
              </>
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

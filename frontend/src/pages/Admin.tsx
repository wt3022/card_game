import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import CardEditor from '../components/CardManagement/CardEditor'
import DeckEditor from '../components/DeckManagement/DeckEditor'
import type { Card, Deck } from '../gen/common_pb'
import { cardManagementClient } from '../lib/api-client'
import { getUserInfo, logout } from '../lib/auth'
import './Admin.css'

type AdminView = 'card-list' | 'card-edit' | 'deck-list' | 'deck-edit'

export default function Admin() {
  const [currentView, setCurrentView] = useState<AdminView>('card-list')
  const [cards, setCards] = useState<Card[]>([])
  const [selectedCard, setSelectedCard] = useState<Card | null>(null)
  const [isNewCardMode, setIsNewCardMode] = useState(false)
  const [decks, setDecks] = useState<Deck[]>([])
  const [selectedDeck, setSelectedDeck] = useState<Deck | null>(null)
  const [isNewDeckMode, setIsNewDeckMode] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sortColumn, setSortColumn] = useState<string>('id')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')
  const [filterText, setFilterText] = useState('')
  const userInfo = getUserInfo()

  const loadCards = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await cardManagementClient.listCards({})
      setCards(response.cards || [])
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'カードの読み込みに失敗しました',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  const loadDecks = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await cardManagementClient.listDecks({})
      setDecks(response.decks || [])
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'デッキの読み込みに失敗しました',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (currentView === 'card-list' || currentView === 'card-edit') {
      loadCards()
    } else if (currentView === 'deck-list' || currentView === 'deck-edit') {
      loadDecks()
    }
  }, [currentView, loadCards, loadDecks])

  const handleSort = (column: string) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortColumn(column)
      setSortDirection('asc')
    }
  }

  const getSortedAndFilteredCards = () => {
    const filtered = cards.filter((card) => {
      if (!filterText) return true
      const searchText = filterText.toLowerCase()
      return (
        card.id.toLowerCase().includes(searchText) ||
        card.name.toLowerCase().includes(searchText) ||
        card.type.toString().toLowerCase().includes(searchText)
      )
    })

    return filtered.sort((a, b) => {
      let aValue: unknown = a[sortColumn as keyof Card]
      let bValue: unknown = b[sortColumn as keyof Card]

      if (
        sortColumn === 'cost' ||
        sortColumn === 'attack' ||
        sortColumn === 'defense'
      ) {
        aValue = (aValue as number | undefined) ?? -1
        bValue = (bValue as number | undefined) ?? -1
      }

      if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1
      if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1
      return 0
    })
  }

  const getSortedAndFilteredDecks = () => {
    const filtered = decks.filter((deck) => {
      if (!filterText) return true
      const searchText = filterText.toLowerCase()
      return (
        deck.id.toLowerCase().includes(searchText) ||
        deck.name.toLowerCase().includes(searchText)
      )
    })

    return filtered.sort((a, b) => {
      const aValue: unknown = a[sortColumn as keyof Deck]
      const bValue: unknown = b[sortColumn as keyof Deck]

      if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1
      if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1
      return 0
    })
  }

  const handleCardSelect = (card: Card) => {
    setSelectedCard(card)
    setIsNewCardMode(false)
    setCurrentView('card-edit')
  }

  const handleNewCardClick = () => {
    setSelectedCard(null)
    setIsNewCardMode(true)
    setCurrentView('card-edit')
  }

  const handleCardSave = async (_savedCardId: string) => {
    await loadCards()
    setIsNewCardMode(false)
    setCurrentView('card-list')
  }

  const handleCardDelete = async (cardId: string) => {
    if (!confirm('このカードを削除しますか？')) {
      return
    }

    try {
      await cardManagementClient.deleteCard({ id: cardId })
      await loadCards()
      if (selectedCard?.id === cardId) {
        setSelectedCard(null)
      }
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'カードの削除に失敗しました')
    }
  }

  const handleDeckSelect = (deck: Deck) => {
    setSelectedDeck(deck)
    setIsNewDeckMode(false)
    setCurrentView('deck-edit')
  }

  const handleNewDeckClick = () => {
    setSelectedDeck(null)
    setIsNewDeckMode(true)
    setCurrentView('deck-edit')
  }

  const handleDeckSave = async (_savedDeckId?: string) => {
    await loadDecks()
    setIsNewDeckMode(false)
    setCurrentView('deck-list')
  }

  const handleDeckDelete = async (deckId: string) => {
    if (!confirm('このデッキを削除しますか？')) {
      return
    }

    try {
      await cardManagementClient.deleteDeck({ id: deckId })
      await loadDecks()
      if (selectedDeck?.id === deckId) {
        setSelectedDeck(null)
      }
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'デッキの削除に失敗しました')
    }
  }

  const handleLogout = () => {
    logout()
    window.location.href = '/'
  }

  return (
    <div className="admin-container">
      <header className="admin-header">
        <div className="admin-header-left">
          <h1>管理画面</h1>
          <Link to="/" className="game-link">
            ゲームに戻る
          </Link>
        </div>
        <div className="admin-header-actions">
          <span className="user-info">
            {userInfo?.username} ({userInfo?.role})
          </span>
          <button type="button" onClick={handleLogout}>
            ログアウト
          </button>
        </div>
      </header>

      {error && <div className="error-banner">{error}</div>}

      <div className="admin-layout">
        <aside className="admin-sidebar">
          <nav className="admin-nav">
            <button
              type="button"
              className={`nav-item ${currentView === 'card-list' || currentView === 'card-edit' ? 'active' : ''}`}
              onClick={() => setCurrentView('card-list')}
            >
              カード管理
            </button>
            <button
              type="button"
              className={`nav-item ${currentView === 'deck-list' || currentView === 'deck-edit' ? 'active' : ''}`}
              onClick={() => setCurrentView('deck-list')}
            >
              デッキ管理
            </button>
          </nav>
        </aside>

        <main className="admin-main-content">
          {currentView === 'card-list' && (
            <div className="list-view">
              <div className="list-header">
                <h2>カード一覧</h2>
                <div className="list-actions">
                  <input
                    type="text"
                    className="filter-input"
                    placeholder="ID、名前、タイプで検索..."
                    value={filterText}
                    onChange={(e) => setFilterText(e.target.value)}
                  />
                  <button
                    type="button"
                    className="btn-primary"
                    onClick={handleNewCardClick}
                  >
                    + 新しいカード
                  </button>
                </div>
              </div>

              {loading ? (
                <div className="loading">読み込み中...</div>
              ) : (
                <div className="table-container">
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th
                          onClick={() => handleSort('id')}
                          className="sortable"
                        >
                          ID{' '}
                          {sortColumn === 'id' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th
                          onClick={() => handleSort('name')}
                          className="sortable"
                        >
                          名前{' '}
                          {sortColumn === 'name' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th
                          onClick={() => handleSort('type')}
                          className="sortable"
                        >
                          タイプ{' '}
                          {sortColumn === 'type' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th
                          onClick={() => handleSort('cost')}
                          className="sortable"
                        >
                          コスト{' '}
                          {sortColumn === 'cost' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th
                          onClick={() => handleSort('attack')}
                          className="sortable"
                        >
                          攻撃力{' '}
                          {sortColumn === 'attack' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th
                          onClick={() => handleSort('defense')}
                          className="sortable"
                        >
                          体力{' '}
                          {sortColumn === 'defense' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th>効果</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {getSortedAndFilteredCards().map((card) => (
                        <tr
                          key={card.id}
                          onClick={() => handleCardSelect(card)}
                          className="clickable"
                        >
                          <td>{card.id}</td>
                          <td>{card.name}</td>
                          <td>{card.type}</td>
                          <td>{card.cost}</td>
                          <td>{card.attack ?? '-'}</td>
                          <td>{card.defense ?? '-'}</td>
                          <td className="effect-cell">{card.effect}</td>
                          <td>
                            <button
                              type="button"
                              className="btn-danger-small"
                              onClick={(e) => {
                                e.stopPropagation()
                                handleCardDelete(card.id)
                              }}
                            >
                              削除
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {currentView === 'card-edit' && (
            <div className="edit-view">
              <div className="edit-header">
                <button
                  type="button"
                  className="btn-back"
                  onClick={() => setCurrentView('card-list')}
                >
                  ← 一覧に戻る
                </button>
                <h2>{isNewCardMode ? '新しいカード' : 'カード編集'}</h2>
              </div>
              <CardEditor
                card={selectedCard}
                isNewCardMode={isNewCardMode}
                onSave={handleCardSave}
                onCancel={() => setCurrentView('card-list')}
              />
            </div>
          )}

          {currentView === 'deck-list' && (
            <div className="list-view">
              <div className="list-header">
                <h2>デッキ一覧</h2>
                <div className="list-actions">
                  <input
                    type="text"
                    className="filter-input"
                    placeholder="ID、名前で検索..."
                    value={filterText}
                    onChange={(e) => setFilterText(e.target.value)}
                  />
                  <button
                    type="button"
                    className="btn-primary"
                    onClick={handleNewDeckClick}
                  >
                    + 新しいデッキ
                  </button>
                </div>
              </div>

              {loading ? (
                <div className="loading">読み込み中...</div>
              ) : (
                <div className="table-container">
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th
                          onClick={() => handleSort('id')}
                          className="sortable"
                        >
                          ID{' '}
                          {sortColumn === 'id' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th
                          onClick={() => handleSort('name')}
                          className="sortable"
                        >
                          名前{' '}
                          {sortColumn === 'name' &&
                            (sortDirection === 'asc' ? '▲' : '▼')}
                        </th>
                        <th>カード枚数</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {getSortedAndFilteredDecks().map((deck) => (
                        <tr
                          key={deck.id}
                          onClick={() => handleDeckSelect(deck)}
                          className="clickable"
                        >
                          <td>{deck.id}</td>
                          <td>{deck.name}</td>
                          <td>{deck.cardIds?.length ?? 0}枚</td>
                          <td>
                            <button
                              type="button"
                              className="btn-danger-small"
                              onClick={(e) => {
                                e.stopPropagation()
                                handleDeckDelete(deck.id)
                              }}
                            >
                              削除
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {currentView === 'deck-edit' && (
            <div className="edit-view">
              <div className="edit-header">
                <button
                  type="button"
                  className="btn-back"
                  onClick={() => setCurrentView('deck-list')}
                >
                  ← 一覧に戻る
                </button>
                <h2>{isNewDeckMode ? '新しいデッキ' : 'デッキ編集'}</h2>
              </div>
              <DeckEditor
                deck={selectedDeck}
                isNewDeckMode={isNewDeckMode}
                onSave={handleDeckSave}
                onCancel={() => setCurrentView('deck-list')}
              />
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

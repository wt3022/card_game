import { useCallback, useEffect, useState } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'
import CardEditor from '../components/CardManagement/CardEditor'
import DeckEditor from '../components/DeckManagement/DeckEditor'
import type { Card, Deck } from '../gen/common_pb'
import { CardType } from '../gen/common_pb'
import { cardManagementClient } from '../lib/api-client'
import { getUserInfo, logout } from '../lib/auth'
import './Admin.css'

type AdminView = 'card-list' | 'card-edit' | 'deck-list' | 'deck-edit'

const getCardTypeLabel = (type: CardType): string => {
  switch (type) {
    case CardType.UNIT:
      return 'ユニット'
    case CardType.SPELL:
      return 'スペル'
    case CardType.LEADER:
      return 'リーダー'
    default:
      return '不明'
  }
}

export default function Admin() {
  const navigate = useNavigate()
  const location = useLocation()
  const { id } = useParams<{ id: string }>()
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
  const [searchInput, setSearchInput] = useState('')
  const [typeFilter, setTypeFilter] = useState<CardType | 'ALL'>('ALL')
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [totalCount, setTotalCount] = useState(0)
  const [pageSize, setPageSize] = useState(50)
  const [isEditMode, setIsEditMode] = useState(false)
  const [deckSearchInput, setDeckSearchInput] = useState('')
  const [isDeckEditMode, setIsDeckEditMode] = useState(false)
  const userInfo = getUserInfo()

  const loadCards = useCallback(
    async (page = 1, size = pageSize) => {
      try {
        setLoading(true)
        setError(null)

        const response = await cardManagementClient.listCards({
          page,
          pageSize: size,
        })
        setCards(response.cards || [])
        setTotalPages(response.totalPages)
        setTotalCount(response.totalCount)
        setCurrentPage(page)
      } catch (err: unknown) {
        console.error('Failed to load cards:', err)
        setError(
          err instanceof Error ? err.message : 'カードの読み込みに失敗しました',
        )
      } finally {
        setLoading(false)
      }
    },
    [pageSize],
  )

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

  const loadCardDetail = useCallback(async (cardId: string) => {
    try {
      setLoading(true)
      setError(null)
      const response = await cardManagementClient.getCard({ id: cardId })
      console.log('Loaded card detail:', response.card)
      console.log('Card effect:', response.card?.cardEffect)
      setSelectedCard(response.card ?? null)
    } catch (err: unknown) {
      console.error('Failed to load card details:', err)
      setError(
        err instanceof Error
          ? err.message
          : 'カード詳細の読み込みに失敗しました',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  const loadDeckDetail = useCallback(async (deckId: string) => {
    try {
      setLoading(true)
      setError(null)
      const response = await cardManagementClient.getDeck({ id: deckId })
      setSelectedDeck(response.deck ?? null)
    } catch (err: unknown) {
      console.error('Failed to load deck details:', err)
      setError(
        err instanceof Error
          ? err.message
          : 'デッキ詳細の読み込みに失敗しました',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  // URLパラメータに基づいて現在のビューを判定
  useEffect(() => {
    const pathname = location.pathname
    console.log('URL changed:', pathname, 'ID:', id)

    if (pathname.startsWith('/admin/cards/')) {
      setCurrentView('card-edit')
      setIsNewCardMode(id === 'new')
      if (id && id !== 'new') {
        loadCardDetail(id)
      } else if (id === 'new') {
        setSelectedCard(null)
      }
    } else if (pathname.startsWith('/admin/decks/')) {
      setCurrentView('deck-edit')
      setIsNewDeckMode(id === 'new')
      if (id && id !== 'new') {
        loadDeckDetail(id)
      } else if (id === 'new') {
        setSelectedDeck(null)
      }
    } else if (pathname === '/admin' || pathname === '/admin/cards') {
      setCurrentView('card-list')
      setSelectedCard(null)
      setIsNewCardMode(false)
    } else if (pathname === '/admin/decks') {
      setCurrentView('deck-list')
      setSelectedDeck(null)
      setIsNewDeckMode(false)
    }
  }, [location.pathname, id, loadCardDetail, loadDeckDetail])

  useEffect(() => {
    if (currentView === 'card-list' || currentView === 'card-edit') {
      loadCards()
    } else if (currentView === 'deck-list' || currentView === 'deck-edit') {
      loadDecks()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentView, loadCards, loadDecks])

  const handleSort = (column: string) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortColumn(column)
      setSortDirection('asc')
    }
  }

  const handleSearch = () => {
    setFilterText(searchInput)
  }

  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  const handleDeckSearch = () => {
    setFilterText(deckSearchInput)
  }

  const handleDeckSearchKeyDown = (
    e: React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if (e.key === 'Enter') {
      handleDeckSearch()
    }
  }

  const getSortedAndFilteredCards = () => {
    const filtered = cards.filter((card) => {
      // タイプフィルター
      if (typeFilter !== 'ALL' && card.type !== typeFilter) {
        return false
      }

      // テキスト検索
      if (!filterText) return true
      const searchText = filterText.toLowerCase()
      return (
        card.id.toLowerCase().includes(searchText) ||
        card.name.toLowerCase().includes(searchText) ||
        card.type.toString().toLowerCase().includes(searchText)
      )
    })

    return filtered.sort((a, b) => {
      const aValue: unknown = a[sortColumn as keyof Card]
      const bValue: unknown = b[sortColumn as keyof Card]

      // 数値フィールドの処理
      if (
        sortColumn === 'cost' ||
        sortColumn === 'attack' ||
        sortColumn === 'defense'
      ) {
        const aNum = (aValue as number | undefined) ?? -1
        const bNum = (bValue as number | undefined) ?? -1

        console.log(
          `Sorting ${sortColumn}: ${a.id}(${aNum}) vs ${b.id}(${bNum}), direction: ${sortDirection}`,
        )

        if (aNum !== bNum) {
          return sortDirection === 'asc' ? aNum - bNum : bNum - aNum
        }
        // 数値が同じ場合はIDで二次ソート
        return a.id.localeCompare(b.id)
      }

      // 文字列フィールドの処理
      const aStr = String(aValue ?? '')
      const bStr = String(bValue ?? '')
      const compareResult = aStr.localeCompare(bStr)

      if (compareResult !== 0) {
        return sortDirection === 'asc' ? compareResult : -compareResult
      }
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

      const aComp = aValue as string | number
      const bComp = bValue as string | number
      if (aComp < bComp) return sortDirection === 'asc' ? -1 : 1
      if (aComp > bComp) return sortDirection === 'asc' ? 1 : -1
      return 0
    })
  }

  const handleCardSelect = (card: Card) => {
    navigate(`/admin/cards/${card.id}`)
  }

  const handleNewCardClick = () => {
    navigate('/admin/cards/new')
  }

  const handleCardSave = async (_savedCardId: string) => {
    await loadCards()
    navigate('/admin')
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
        navigate('/admin')
      }
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'カードの削除に失敗しました')
    }
  }

  const handleDeckSelect = (deck: Deck) => {
    navigate(`/admin/decks/${deck.id}`)
  }

  const handleNewDeckClick = () => {
    navigate('/admin/decks/new')
  }

  const handleDeckSave = async (_savedDeckId?: string) => {
    await loadDecks()
    navigate('/admin/decks')
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
        navigate('/admin/decks')
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
              onClick={() => navigate('/admin')}
            >
              カード管理
            </button>
            <button
              type="button"
              className={`nav-item ${currentView === 'deck-list' || currentView === 'deck-edit' ? 'active' : ''}`}
              onClick={() => navigate('/admin/decks')}
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
                  <button
                    type="button"
                    className="btn-primary"
                    onClick={handleNewCardClick}
                  >
                    + 新しいカード
                  </button>
                  <button
                    type="button"
                    className={`btn-edit-mode ${isEditMode ? 'active' : ''}`}
                    onClick={() => setIsEditMode(!isEditMode)}
                  >
                    {isEditMode ? '✓ 編集モード' : '編集モード'}
                  </button>
                </div>
              </div>

              {/* フィルター・ページネーションコントロール */}
              <div className="list-controls">
                <div className="filter-controls">
                  <div className="type-filter-group">
                    <span className="filter-label">種別:</span>
                    <button
                      type="button"
                      className={`type-filter-btn ${typeFilter === 'ALL' ? 'active' : ''}`}
                      onClick={() => setTypeFilter('ALL')}
                    >
                      すべて
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
                    <button
                      type="button"
                      className={`type-filter-btn ${typeFilter === CardType.LEADER ? 'active' : ''}`}
                      onClick={() => setTypeFilter(CardType.LEADER)}
                    >
                      リーダー
                    </button>
                  </div>
                  <div className="search-box">
                    <input
                      type="text"
                      className="filter-input"
                      placeholder="ID、名前で検索..."
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      onKeyDown={handleSearchKeyDown}
                    />
                    <button
                      type="button"
                      className="btn-search"
                      onClick={handleSearch}
                    >
                      検索
                    </button>
                  </div>
                </div>
                <div className="page-size-selector">
                  <label htmlFor="page-size-top">表示件数:</label>
                  <select
                    id="page-size-top"
                    value={pageSize}
                    onChange={(e) => {
                      const newSize = Number(e.target.value)
                      setPageSize(newSize)
                      setCurrentPage(1)
                      loadCards(1, newSize)
                    }}
                    className="page-size-select"
                  >
                    <option value={30}>30件</option>
                    <option value={50}>50件</option>
                    <option value={100}>100件</option>
                  </select>
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
                        {isEditMode && <th>操作</th>}
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
                          <td>{getCardTypeLabel(card.type)}</td>
                          <td>{card.cost}</td>
                          <td>{card.attack ?? '-'}</td>
                          <td>{card.defense ?? '-'}</td>
                          <td className="effect-cell">{card.effect}</td>
                          {isEditMode && (
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
                          )}
                        </tr>
                      ))}
                    </tbody>
                  </table>

                  {/* ページネーション */}
                  {totalCount > 0 && (
                    <div className="table-pagination">
                      <button
                        type="button"
                        onClick={() =>
                          loadCards(Math.max(1, currentPage - 1), pageSize)
                        }
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
                          loadCards(
                            Math.min(totalPages, currentPage + 1),
                            pageSize,
                          )
                        }
                        disabled={currentPage >= totalPages}
                        className="pagination-btn"
                      >
                        次へ →
                      </button>
                    </div>
                  )}
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
                  onClick={() => navigate('/admin')}
                >
                  ← 一覧に戻る
                </button>
                <h2>{isNewCardMode ? '新しいカード' : 'カード編集'}</h2>
              </div>
              <CardEditor
                card={selectedCard}
                isNewCardMode={isNewCardMode}
                onSave={handleCardSave}
                onCancel={() => navigate('/admin')}
              />
            </div>
          )}

          {currentView === 'deck-list' && (
            <div className="list-view">
              <div className="list-header">
                <h2>デッキ一覧</h2>
                <div className="list-actions">
                  <button
                    type="button"
                    className="btn-primary"
                    onClick={handleNewDeckClick}
                  >
                    + 新しいデッキ
                  </button>
                  <button
                    type="button"
                    className={`btn-edit-mode ${isDeckEditMode ? 'active' : ''}`}
                    onClick={() => setIsDeckEditMode(!isDeckEditMode)}
                  >
                    {isDeckEditMode ? '✓ 編集モード' : '編集モード'}
                  </button>
                </div>
              </div>

              {/* 検索コントロール */}
              <div className="list-controls">
                <div className="filter-controls">
                  <div className="search-box">
                    <input
                      type="text"
                      className="filter-input"
                      placeholder="ID、名前で検索..."
                      value={deckSearchInput}
                      onChange={(e) => setDeckSearchInput(e.target.value)}
                      onKeyDown={handleDeckSearchKeyDown}
                    />
                    <button
                      type="button"
                      className="btn-search"
                      onClick={handleDeckSearch}
                    >
                      検索
                    </button>
                  </div>
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
                        {isDeckEditMode && <th>操作</th>}
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
                          {isDeckEditMode && (
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
                          )}
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
                  onClick={() => navigate('/admin/decks')}
                >
                  ← 一覧に戻る
                </button>
                <h2>{isNewDeckMode ? '新しいデッキ' : 'デッキ編集'}</h2>
              </div>
              <DeckEditor
                deck={selectedDeck}
                isNewDeckMode={isNewDeckMode}
                onSave={handleDeckSave}
                onCancel={() => navigate('/admin/decks')}
              />
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

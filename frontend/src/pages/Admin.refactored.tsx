import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import CardEditor from '../components/CardManagement/CardEditor'
import CardList from '../components/CardManagement/CardList'
import DeckEditor from '../components/DeckManagement/DeckEditor'
import DeckList from '../components/DeckManagement/DeckList'
import type { Card, Deck } from '../gen/common_pb'
import { cardManagementClient } from '../lib/api-client'
import { getUserInfo, logout } from '../lib/auth'
import './Admin.css'

type AdminTab = 'cards' | 'decks'

export default function Admin() {
  const [activeTab, setActiveTab] = useState<AdminTab>('cards')
  const [cards, setCards] = useState<Card[]>([])
  const [selectedCard, setSelectedCard] = useState<Card | null>(null)
  const [isNewCardMode, setIsNewCardMode] = useState(false)
  const [decks, setDecks] = useState<Deck[]>([])
  const [selectedDeck, setSelectedDeck] = useState<Deck | null>(null)
  const [isNewDeckMode, setIsNewDeckMode] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
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
    if (activeTab === 'cards') {
      loadCards()
    } else if (activeTab === 'decks') {
      loadDecks()
    }
  }, [activeTab, loadCards, loadDecks])

  const handleCardSelect = (card: Card) => {
    setSelectedCard(card)
    setIsNewCardMode(false)
  }

  const handleNewCardClick = () => {
    setSelectedCard(null)
    setIsNewCardMode(true)
  }

  const handleCardSave = async () => {
    await loadCards()
    setSelectedCard(null)
    setIsNewCardMode(false)
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
  }

  const handleNewDeckClick = () => {
    setSelectedDeck(null)
    setIsNewDeckMode(true)
  }

  const handleDeckSave = async () => {
    await loadDecks()
    setSelectedDeck(null)
    setIsNewDeckMode(false)
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

      <div className="admin-tabs">
        <button
          type="button"
          className={`tab-button ${activeTab === 'cards' ? 'active' : ''}`}
          onClick={() => setActiveTab('cards')}
        >
          カード管理
        </button>
        <button
          type="button"
          className={`tab-button ${activeTab === 'decks' ? 'active' : ''}`}
          onClick={() => setActiveTab('decks')}
        >
          デッキ管理
        </button>
      </div>

      {error && <div className="error-banner">{error}</div>}

      {activeTab === 'cards' && (
        <div className="admin-content">
          <div className="admin-sidebar">
            <button
              type="button"
              className="new-card-button"
              onClick={handleNewCardClick}
            >
              + 新しいカード
            </button>
            <CardList
              cards={cards}
              selectedCard={selectedCard}
              onCardSelect={handleCardSelect}
              onCardDelete={handleCardDelete}
              loading={loading}
            />
          </div>
          <div className="admin-main">
            <CardEditor
              card={selectedCard}
              isNewCardMode={isNewCardMode}
              onSave={handleCardSave}
              onCancel={() => {
                setSelectedCard(null)
                setIsNewCardMode(false)
              }}
            />
          </div>
        </div>
      )}

      {activeTab === 'decks' && (
        <div className="admin-content">
          <div className="admin-sidebar">
            <button
              type="button"
              className="new-deck-button"
              onClick={handleNewDeckClick}
            >
              + 新しいデッキ
            </button>
            <DeckList
              decks={decks}
              selectedDeck={selectedDeck}
              onDeckSelect={handleDeckSelect}
              onDeckDelete={handleDeckDelete}
              loading={loading}
            />
          </div>
          <div className="admin-main">
            <DeckEditor
              deck={selectedDeck}
              isNewDeckMode={isNewDeckMode}
              onSave={handleDeckSave}
              onCancel={() => {
                setSelectedDeck(null)
                setIsNewDeckMode(false)
              }}
            />
          </div>
        </div>
      )}
    </div>
  )
}

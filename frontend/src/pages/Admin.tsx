import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import CardEditor from '../components/CardManagement/CardEditor'
import CardList from '../components/CardManagement/CardList'
import { CardManagementService } from '../gen/card_management_connect'
import type { Card } from '../gen/common_pb'
import { createAuthenticatedClient, getUserInfo, logout } from '../lib/auth'
import './Admin.css'

export default function Admin() {
  const [cards, setCards] = useState<Card[]>([])
  const [selectedCard, setSelectedCard] = useState<Card | null>(null)
  const [isNewCardMode, setIsNewCardMode] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const userInfo = getUserInfo()

  const cardClient = useMemo(
    () => createAuthenticatedClient(CardManagementService),
    [],
  )

  const loadCards = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await cardClient.listCards({})
      setCards(response.cards || [])
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'カードの読み込みに失敗しました',
      )
    } finally {
      setLoading(false)
    }
  }, [cardClient])

  useEffect(() => {
    loadCards()
  }, [loadCards])

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
      await cardClient.deleteCard({ id: cardId })
      await loadCards()
      if (selectedCard?.id === cardId) {
        setSelectedCard(null)
      }
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'カードの削除に失敗しました')
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
          <h1>カード管理</h1>
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
    </div>
  )
}

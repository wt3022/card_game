import type { Card } from '../../gen/common_pb'
import './CardList.css'

interface CardListProps {
  cards: Card[]
  selectedCard: Card | null
  onCardSelect: (card: Card) => void
  onCardDelete: (cardId: string) => void
  loading: boolean
}

export default function CardList({
  cards,
  selectedCard,
  onCardSelect,
  onCardDelete,
  loading,
}: CardListProps) {
  if (loading) {
    return <div className="card-list-loading">読み込み中...</div>
  }

  if (cards.length === 0) {
    return <div className="card-list-empty">カードがありません</div>
  }

  return (
    <div className="card-list">
      {cards.map((card) => (
        <div
          key={card.id}
          className={`card-item ${
            selectedCard?.id === card.id ? 'selected' : ''
          }`}
          role="button"
          tabIndex={0}
          onClick={() => onCardSelect(card)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              onCardSelect(card)
            }
          }}
        >
          <div className="card-item-header">
            <span className="card-item-name">{card.name}</span>
            <span className="card-item-cost">{card.cost}</span>
          </div>
          <div className="card-item-type">{card.type}</div>
          {card.attack !== undefined && card.defense !== undefined && (
            <div className="card-item-stats">
              {card.attack}/{card.defense}
            </div>
          )}
          <button
            type="button"
            className="card-item-delete"
            onClick={(e) => {
              e.stopPropagation()
              onCardDelete(card.id)
            }}
          >
            削除
          </button>
        </div>
      ))}
    </div>
  )
}

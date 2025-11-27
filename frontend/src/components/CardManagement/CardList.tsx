import type { Card } from '../../gen/common_pb'
import './CardList.css'

interface CardListProps {
  cards: Card[]
  selectedCard: Card | null
  onCardSelect: (card: Card) => void
  onCardDelete: (cardId: string) => void
  loading: boolean
  currentPage?: number
  totalPages?: number
  totalCount?: number
  pageSize?: number
  onPageChange?: (page: number, pageSize: number) => void
  onPageSizeChange?: (pageSize: number) => void
}

export default function CardList({
  cards,
  selectedCard,
  onCardSelect,
  onCardDelete,
  loading,
  currentPage = 1,
  totalPages = 1,
  totalCount = 0,
  pageSize = 50,
  onPageChange,
  onPageSizeChange,
}: CardListProps) {
  if (loading) {
    return <div className="card-list-loading">読み込み中...</div>
  }

  if (cards.length === 0) {
    return <div className="card-list-empty">カードがありません</div>
  }

  return (
    <div className="card-list-container">
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

      {onPageChange && totalCount > 0 && (
        <div className="card-list-pagination">
          <div className="pagination-controls">
            <button
              type="button"
              onClick={() =>
                onPageChange(Math.max(1, currentPage - 1), pageSize)
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
                onPageChange(Math.min(totalPages, currentPage + 1), pageSize)
              }
              disabled={currentPage >= totalPages}
              className="pagination-btn"
            >
              次へ →
            </button>
          </div>
          {onPageSizeChange && (
            <div className="page-size-selector">
              <label htmlFor="card-page-size">表示件数:</label>
              <select
                id="card-page-size"
                value={pageSize}
                onChange={(e) => onPageSizeChange(Number(e.target.value))}
                className="page-size-select"
              >
                <option value={30}>30件</option>
                <option value={50}>50件</option>
                <option value={100}>100件</option>
              </select>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

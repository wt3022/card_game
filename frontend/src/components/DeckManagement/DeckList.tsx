import type { Deck } from '../../gen/common_pb'
import './DeckList.css'

interface DeckListProps {
  decks: Deck[]
  selectedDeck: Deck | null
  onDeckSelect: (deck: Deck) => void
  onDeckDelete: (deckId: string) => void
  loading: boolean
}

export default function DeckList({
  decks,
  selectedDeck,
  onDeckSelect,
  onDeckDelete,
  loading,
}: DeckListProps) {
  if (loading) {
    return <div className="deck-loading">読み込み中...</div>
  }

  if (decks.length === 0) {
    return <div className="deck-empty">デッキがありません</div>
  }

  return (
    <div className="deck-list">
      {decks.map((deck) => (
        <div
          key={deck.id}
          className={`deck-item ${selectedDeck?.id === deck.id ? 'selected' : ''}`}
          onClick={() => onDeckSelect(deck)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              onDeckSelect(deck)
            }
          }}
          role="button"
          tabIndex={0}
        >
          <div className="deck-item-header">
            <span className="deck-item-name">{deck.name}</span>
            <div className="deck-item-actions">
              <button
                type="button"
                className="deck-item-delete"
                onClick={(e) => {
                  e.stopPropagation()
                  onDeckDelete(deck.id)
                }}
              >
                削除
              </button>
            </div>
          </div>
          <div className="deck-item-info">
            <span>{deck.cardIds.length} 枚</span>
            {deck.description && <span>{deck.description}</span>}
          </div>
        </div>
      ))}
    </div>
  )
}

import { useState, useEffect } from 'react'
import type { Card } from '../gen/common_pb'
import './MulliganModal.css'

interface MulliganModalProps {
  hand: Card[]
  onSubmit: (selectedCardIds: string[]) => void
  onSkip: () => void
  isWaitingForOpponent: boolean
}

export default function MulliganModal({
  hand,
  onSubmit,
  onSkip,
  isWaitingForOpponent,
}: MulliganModalProps) {
  const [selectedCardIds, setSelectedCardIds] = useState<Set<string>>(new Set())

  const toggleCardSelection = (cardId: string) => {
    const newSelection = new Set(selectedCardIds)
    if (newSelection.has(cardId)) {
      newSelection.delete(cardId)
    } else {
      newSelection.add(cardId)
    }
    setSelectedCardIds(newSelection)
  }

  const handleSubmit = () => {
    onSubmit(Array.from(selectedCardIds))
  }

  // 手札が空の場合は自動的にスキップ
  useEffect(() => {
    if (hand.length === 0 && !isWaitingForOpponent) {
      console.log('手札が空のため、自動的にマリガンをスキップします')
      const timer = setTimeout(() => onSkip(), 100)
      return () => clearTimeout(timer)
    }
  }, [hand.length, isWaitingForOpponent, onSkip])

  if (isWaitingForOpponent) {
    return (
      <div className="mulligan-modal-overlay">
        <div className="mulligan-modal">
          <h2>マリガン完了</h2>
          <p>相手プレイヤーのマリガンを待っています...</p>
          <div className="loading-spinner"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="mulligan-modal-overlay">
      <div className="mulligan-modal">
        <h2>マリガン</h2>
        <p>引き直したいカードを選択してください（{selectedCardIds.size}枚選択中）</p>
        
        <div className="mulligan-cards">
          {hand.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px', color: '#fff' }}>
              手札がありません。スキップしてください。
            </div>
          ) : (
            hand.map((card) => (
            <div
              key={card.id}
              className={`mulligan-card ${selectedCardIds.has(card.id) ? 'selected' : ''}`}
              onClick={() => toggleCardSelection(card.id)}
            >
              <div className="card-cost">{card.cost}</div>
              <div className="card-name">{card.name}</div>
              {card.effect && <div className="card-effect">{card.effect}</div>}
              {card.attack !== undefined && card.attack !== null && (
                <div className="card-stats">
                  <span className="atk">⚔️{card.attack}</span>
                  <span className="def">🛡️{card.defense}</span>
                </div>
              )}
              {selectedCardIds.has(card.id) && (
                <div className="selected-badge">✓</div>
              )}
            </div>
            ))
          )}
        </div>

        <div className="mulligan-actions">
          <button
            className="btn-skip"
            onClick={onSkip}
          >
            スキップ
          </button>
          <button
            className="btn-submit"
            onClick={handleSubmit}
            disabled={selectedCardIds.size === 0}
          >
            {selectedCardIds.size > 0 
              ? `${selectedCardIds.size}枚引き直す` 
              : 'カードを選択してください'}
          </button>
        </div>
      </div>
    </div>
  )
}


import './HandCard.css'

interface HandCardProps {
  cardIndex: number
  isSelected: boolean
  onClick: () => void
}

export default function HandCard({ cardIndex, isSelected, onClick }: HandCardProps) {
  return (
    <div
      className={`hand-card ${isSelected ? 'selected' : ''}`}
      onClick={onClick}
    >
      <div className="card-back">
        <div className="card-number">#{cardIndex + 1}</div>
        <div className="card-pattern">🎴</div>
      </div>
    </div>
  )
}


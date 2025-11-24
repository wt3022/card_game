import type { Player } from '../gen/common_pb'
import './PlayerInfo.css'

interface PlayerInfoProps {
  player: Player | undefined
  isCurrentPlayer: boolean
  onClick?: () => void
  isAttackTarget?: boolean
}

export default function PlayerInfo({
  player,
  isCurrentPlayer,
  onClick,
  isAttackTarget = false,
}: PlayerInfoProps) {
  if (!player) {
    return (
      <div className="player-info">
        <div className="player-name">プレイヤー情報を読み込み中...</div>
      </div>
    )
  }

  const interactiveProps = onClick
    ? {
        role: 'button' as const,
        tabIndex: 0,
        onClick,
        onKeyDown: (e: React.KeyboardEvent) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onClick()
          }
        },
      }
    : {}

  return (
    <div
      className={`player-info-compact ${isCurrentPlayer ? 'current' : ''} ${isAttackTarget ? 'attack-target' : ''}`}
      {...interactiveProps}
      style={{ cursor: isAttackTarget ? 'pointer' : 'default' }}
    >
      <div className="player-name-compact">{player.name}</div>
      <div className="player-stats-row">
        <div className="stat-compact hp">
          <div className="stat-label">HP</div>
          <div className="stat-value">
            {player.hp}/{player.maxHp}
          </div>
        </div>
        <div className="stat-compact mana">
          <div className="stat-label">マナ</div>
          <div className="stat-value">
            {player.currentTurnMana}/{player.currentRecoveryMana}
          </div>
        </div>
        <div className="stat-compact deck">
          <div className="stat-label">山札</div>
          <div className="stat-value">{player.deckCount}</div>
        </div>
        <div className="stat-compact hand">
          <div className="stat-label">手札</div>
          <div className="stat-value">{player.handCount}</div>
        </div>
        <div className="stat-compact graveyard">
          <div className="stat-label">墓地</div>
          <div className="stat-value">{player.graveyardCount}</div>
        </div>
      </div>
    </div>
  )
}

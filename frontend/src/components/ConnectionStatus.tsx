import type { Player } from '../gen/common_pb'
import './ConnectionStatus.css'

interface ConnectionStatusProps {
  opponent: Player | undefined
}

export default function ConnectionStatus({ opponent }: ConnectionStatusProps) {
  if (!opponent || opponent.isConnected) {
    return null
  }

  return (
    <div className="connection-status-banner">
      <span className="warning-icon">⚠️</span>
      <span>{opponent.name}が退席しました</span>
    </div>
  )
}

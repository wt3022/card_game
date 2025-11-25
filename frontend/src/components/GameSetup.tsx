import { useState } from 'react'
import type { GameState } from '../gen/common_pb'
import { useMatchmaking } from '../hooks/useMatchmaking'
import { useMulligan } from '../hooks/useMulligan'
import type { GameSetupProps } from '../types/components'
import { getMulliganHand } from '../utils/gameHelpers'
import MulliganModal from './MulliganModal'
import './GameSetup.css'

export default function GameSetup({ onGameStart }: GameSetupProps) {
  const [playerName, setPlayerName] = useState('')
  const [showMulligan, setShowMulligan] = useState(false)
  const [mulliganGameState, setMulliganGameState] = useState<GameState | null>(
    null,
  )
  const [mulliganPlayerId, setMulliganPlayerId] = useState('')

  // マッチング管理
  const {
    isMatchmaking,
    matchmakingStatus,
    error: matchmakingError,
    joinMatchmaking,
    cancelMatchmaking,
  } = useMatchmaking()

  // マリガン管理
  const {
    performMulligan,
    isWaitingForOpponent,
    error: mulliganError,
  } = useMulligan()

  // エラーメッセージの統合
  const error = matchmakingError || mulliganError

  /**
   * マッチング開始
   */
  const handleJoinMatchmaking = async () => {
    await joinMatchmaking(playerName, (gameState, playerId) => {
      setMulliganGameState(gameState)
      setMulliganPlayerId(playerId)
      setShowMulligan(true)
    })
  }

  /**
   * マリガン実行
   */
  const handleMulliganSubmit = async (selectedCardIds: string[]) => {
    if (!mulliganGameState || !mulliganPlayerId) return

    await performMulligan(
      mulliganGameState.gameId,
      mulliganPlayerId,
      selectedCardIds,
      (gameState) => {
        onGameStart(gameState, mulliganPlayerId)
      },
    )
  }

  /**
   * マリガンスキップ
   */
  const handleMulliganSkip = async () => {
    await handleMulliganSubmit([])
  }

  /**
   * Enterキーでマッチング開始
   */
  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && playerName.trim()) {
      handleJoinMatchmaking()
    }
  }

  // マリガン画面
  if (showMulligan && mulliganGameState) {
    return (
      <MulliganModal
        hand={getMulliganHand(mulliganGameState, mulliganPlayerId)}
        onSubmit={handleMulliganSubmit}
        onSkip={handleMulliganSkip}
        isWaitingForOpponent={isWaitingForOpponent}
      />
    )
  }

  // マッチング画面
  return (
    <div className="game-setup">
      <div className="setup-card">
        {!isMatchmaking ? (
          <>
            <h2>🎮 対戦相手を探す</h2>
            <p className="description">
              名前を入力してマッチングを開始しましょう
            </p>

            <div className="form-group">
              <label htmlFor="playerName">あなたの名前</label>
              <input
                id="playerName"
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="名前を入力してください"
              />
            </div>

            {error && <div className="error-message">{error}</div>}

            <button
              type="button"
              className="start-button"
              onClick={handleJoinMatchmaking}
              disabled={!playerName.trim()}
            >
              マッチング開始
            </button>
          </>
        ) : (
          <>
            <h2>⏳ マッチング中...</h2>
            <div className="matchmaking-status">
              <div className="spinner"></div>
              <p>{matchmakingStatus}</p>
            </div>

            {error && <div className="error-message">{error}</div>}

            <button
              type="button"
              className="cancel-button"
              onClick={cancelMatchmaking}
            >
              キャンセル
            </button>
          </>
        )}
      </div>
    </div>
  )
}

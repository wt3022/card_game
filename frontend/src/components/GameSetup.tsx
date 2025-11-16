import { useState } from 'react'
import { gameClient } from '../lib/api-client'
import type { GameState } from '../gen/common_pb'
import { MatchmakingStatus } from '../gen/game_pb'
import './GameSetup.css'

interface GameSetupProps {
  onGameStart: (gameState: GameState, playerId: string) => void
}

export default function GameSetup({ onGameStart }: GameSetupProps) {
  const [playerName, setPlayerName] = useState('')
  const [matchmaking, setMatchmaking] = useState(false)
  const [matchmakingStatus, setMatchmakingStatus] = useState<string>('')
  const [error, setError] = useState<string | null>(null)

  const handleJoinMatchmaking = async () => {
    if (!playerName.trim()) {
      setError('名前を入力してください')
      return
    }

    setMatchmaking(true)
    setError(null)
    setMatchmakingStatus('マッチング開始...')

    // ユニークなプレイヤーIDを生成
    const timestamp = Date.now()
    const randomId = Math.random().toString(36).substring(2, 9)
    const playerId = `player-${timestamp}-${randomId}`

    console.log('Joining matchmaking:', { playerId, playerName })

    try {
      // サーバーストリーミングを開始
      const stream = gameClient.joinMatchmaking({
        playerId,
        playerName,
      })

      // ストリームからメッセージを受信
      for await (const response of stream) {
        console.log('Matchmaking response:', response)

        switch (response.status) {
          case MatchmakingStatus.WAITING:
            setMatchmakingStatus(response.message || 'マッチング相手を探しています...')
            break

          case MatchmakingStatus.MATCHED:
            setMatchmakingStatus(response.message || 'マッチング成功！')
            break

          case MatchmakingStatus.GAME_STARTED:
            setMatchmakingStatus('ゲーム開始！')
            if (response.gameState) {
              // ゲーム画面に遷移
              setTimeout(() => {
                onGameStart(response.gameState!, playerId)
              }, 1000)
            }
            return // ストリーム終了

          default:
            setError(response.message || 'マッチングエラー')
            setMatchmaking(false)
            return
        }
      }
    } catch (err) {
      console.error('Matchmaking error:', err)
      setError(err instanceof Error ? err.message : 'マッチングに失敗しました')
      setMatchmaking(false)
    }
  }

  const handleCancel = () => {
    setMatchmaking(false)
    setMatchmakingStatus('')
    setError(null)
  }

  return (
    <div className="game-setup">
      <div className="setup-card">
        {!matchmaking ? (
          <>
            <h2>🎮 対戦相手を探す</h2>
            <p className="description">名前を入力してマッチングを開始しましょう</p>

            <div className="form-group">
              <label htmlFor="playerName">あなたの名前</label>
              <input
                id="playerName"
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && playerName.trim()) {
                    handleJoinMatchmaking()
                  }
                }}
                placeholder="名前を入力してください"
                autoFocus
              />
            </div>

            {error && <div className="error-message">{error}</div>}

            <button
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

            <button className="cancel-button" onClick={handleCancel}>
              キャンセル
            </button>
          </>
        )}
      </div>
    </div>
  )
}


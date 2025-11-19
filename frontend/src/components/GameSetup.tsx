import { useState } from 'react'
import { gameClient } from '../lib/api-client'
import type { GameState } from '../gen/common_pb'
import { MatchmakingStatus } from '../gen/game_pb'
import MulliganModal from './MulliganModal'
import './GameSetup.css'

interface GameSetupProps {
  onGameStart: (gameState: GameState, playerId: string) => void
}

export default function GameSetup({ onGameStart }: GameSetupProps) {
  const [playerName, setPlayerName] = useState('')
  const [matchmaking, setMatchmaking] = useState(false)
  const [matchmakingStatus, setMatchmakingStatus] = useState<string>('')
  const [error, setError] = useState<string | null>(null)
  const [showMulligan, setShowMulligan] = useState(false)
  const [mulliganGameState, setMulliganGameState] = useState<GameState | null>(null)
  const [mulliganPlayerId, setMulliganPlayerId] = useState('')
  const [isWaitingForOpponent, setIsWaitingForOpponent] = useState(false)

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
            setMatchmakingStatus('マリガン中...')
            if (response.gameState) {
              // マリガン画面を表示
              setMulliganGameState(response.gameState)
              setMulliganPlayerId(playerId)
              setShowMulligan(true)
              setMatchmaking(false)
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

  const handleMulliganSubmit = async (selectedCardIds: string[]) => {
    if (!mulliganGameState || !mulliganPlayerId) return

    try {
      // マリガンを実行
      const response = await gameClient.performMulligan({
        gameId: mulliganGameState.gameId,
        playerId: mulliganPlayerId,
        cardIds: selectedCardIds,
      })

      console.log('Mulligan response:', response)

      if (response.success && response.gameState) {
        // 両プレイヤーのマリガンがすでに完了しているかチェック
        const bothDone = response.gameState.player1MulliganDone && response.gameState.player2MulliganDone
        
        if (bothDone) {
          // 両プレイヤーのマリガンが完了している場合、最新の状態を取得してゲーム開始
          console.log('両プレイヤーのマリガンが完了しました。ゲームを開始します。')
          
          // サーバー側でStartTurnが実行されるのを少し待つ
          await new Promise(resolve => setTimeout(resolve, 300))
          
          // 最新のゲーム状態を取得してゲーム開始
          try {
            const latestStateResponse = await gameClient.getGameState({
              gameId: response.gameState.gameId,
              playerId: mulliganPlayerId,
            })
            
            if (latestStateResponse.gameState) {
              console.log('最新のゲーム状態を取得しました。ゲームを開始します。')
              onGameStart(latestStateResponse.gameState, mulliganPlayerId)
            } else {
              // 取得できない場合は現在の状態で開始
              console.log('最新の状態を取得できませんでした。現在の状態で開始します。')
              onGameStart(response.gameState, mulliganPlayerId)
            }
          } catch (err) {
            console.error('GetGameState error:', err)
            // エラーの場合は現在の状態で開始
            onGameStart(response.gameState, mulliganPlayerId)
          }
        } else {
          // まだ相手が完了していない場合、待機状態にしてイベントを購読
          setIsWaitingForOpponent(true)
          subscribeToMulliganCompletion(response.gameState.gameId, mulliganPlayerId)
        }
      } else {
        setError(response.message || 'マリガンに失敗しました')
      }
    } catch (err) {
      console.error('Mulligan error:', err)
      setError(err instanceof Error ? err.message : 'マリガンに失敗しました')
    }
  }

  const handleMulliganSkip = async () => {
    // 空の配列でマリガンを実行（スキップ）
    await handleMulliganSubmit([])
  }

  const subscribeToMulliganCompletion = async (gameId: string, playerId: string) => {
    try {
      const stream = gameClient.streamGameEvents({
        gameId,
        playerId,
      })

      for await (const response of stream) {
        console.log('Game event:', response)

        // マリガン完了イベントを待つ
        if (response.event?.eventType === 'mulligan_completed' && response.gameState) {
          console.log('両プレイヤーのマリガンが完了しました')
          // サーバー側で自動的にStartTurnが実行されるので、次のturn_startedイベントを待つ
        }

        // ターン開始イベントでゲームを開始
        if (response.event?.eventType === 'turn_started' && response.gameState) {
          console.log('ゲーム開始')
          onGameStart(response.gameState, playerId)
          return // ストリーム終了
        }
      }
    } catch (err) {
      console.error('Stream error:', err)
      setError('ゲーム開始に失敗しました')
    }
  }

  const handleCancel = () => {
    setMatchmaking(false)
    setMatchmakingStatus('')
    setError(null)
  }

    // マリガン用の手札を取得
    const getMulliganHand = () => {
      if (!mulliganGameState) return []
      
      // 自分のプレイヤーを特定
      let myHand = []
      if (mulliganGameState.player1?.id === mulliganPlayerId) {
        myHand = mulliganGameState.player1.hand || []
        console.log('プレイヤー1として手札取得:', myHand.length, '枚')
      } else if (mulliganGameState.player2?.id === mulliganPlayerId) {
        myHand = mulliganGameState.player2.hand || []
        console.log('プレイヤー2として手札取得:', myHand.length, '枚')
      } else {
        console.error('プレイヤーIDが一致しません', {
          mulliganPlayerId,
          player1Id: mulliganGameState.player1?.id,
          player2Id: mulliganGameState.player2?.id,
        })
      }
      
      return myHand
    }

    return (
    <div className="game-setup">
      {showMulligan && mulliganGameState ? (
        <MulliganModal
          hand={getMulliganHand()}
          onSubmit={handleMulliganSubmit}
          onSkip={handleMulliganSkip}
          isWaitingForOpponent={isWaitingForOpponent}
        />
      ) : (
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
      )}
    </div>
  )
}

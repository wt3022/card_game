import { useCallback, useEffect, useState } from 'react'
import type { GameState } from '../gen/common_pb'
import { gameClient } from '../lib/api-client'

/**
 * ゲーム状態管理のカスタムフック
 */
export const useGameState = (
  gameId: string | null,
  currentPlayerId: string | null,
) => {
  const [gameState, setGameState] = useState<GameState | null>(null)
  const [error, setError] = useState<string | null>(null)

  /**
   * ゲームイベントのリアルタイム購読
   */
  useEffect(() => {
    console.log(
      'useGameState useEffectがトリガーされました - gameId:',
      gameId,
      'currentPlayerId:',
      currentPlayerId,
    )

    if (!gameId || !currentPlayerId) {
      console.log('❌ 購読をスキップ - gameIdまたはcurrentPlayerIdがありません')
      return
    }

    let abortController: AbortController | null = null
    let isActive = true

    const subscribeToEvents = async () => {
      try {
        console.log(
          '🔌 ゲームイベントを購読中 gameId:',
          gameId,
          'playerId:',
          currentPlayerId,
        )

        abortController = new AbortController()

        const stream = gameClient.streamGameEvents(
          {
            gameId,
            playerId: currentPlayerId,
          },
          { signal: abortController.signal },
        )

        console.log('✅ ストリームが正常に作成されました')

        for await (const response of stream) {
          if (!isActive) break

          console.log('ゲームイベント受信:', response)

          if (response.gameState) {
            console.log(
              'Player1 接続状態:',
              response.gameState.player1?.isConnected,
            )
            console.log(
              'Player2 接続状態:',
              response.gameState.player2?.isConnected,
            )
            setGameState(response.gameState)
          }
        }
      } catch (err: unknown) {
        if (isActive && err instanceof Error && err.name !== 'AbortError') {
          console.error('ストリームエラー:', err)
          setError(err.message)
          // エラー時は再接続を試みる
          setTimeout(() => {
            if (isActive) {
              subscribeToEvents()
            }
          }, 3000)
        }
      }
    }

    subscribeToEvents()

    return () => {
      console.log('🔌 ゲームイベント購読をクリーンアップ中')
      isActive = false
      if (abortController) {
        abortController.abort()
      }
    }
  }, [gameId, currentPlayerId])

  /**
   * ゲーム状態を手動で更新
   */
  const updateGameState = useCallback((newState: GameState) => {
    setGameState(newState)
  }, [])

  /**
   * エラーをクリア
   */
  const clearError = useCallback(() => {
    setError(null)
  }, [])

  return {
    gameState,
    error,
    updateGameState,
    clearError,
  }
}

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
    if (!gameId || !currentPlayerId) return

    let abortController: AbortController | null = null
    let isActive = true

    const subscribeToEvents = async () => {
      try {
        console.log('Subscribing to game events for gameId:', gameId)

        abortController = new AbortController()

        const stream = gameClient.streamGameEvents(
          {
            gameId,
            playerId: currentPlayerId,
          },
          { signal: abortController.signal },
        )

        for await (const response of stream) {
          if (!isActive) break

          console.log('Received game event:', response)

          if (response.gameState) {
            setGameState(response.gameState)
          }
        }
      } catch (err: unknown) {
        if (isActive && err instanceof Error && err.name !== 'AbortError') {
          console.error('Stream error:', err)
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

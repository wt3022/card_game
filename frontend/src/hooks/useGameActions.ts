import { useCallback, useState } from 'react'
import type { GameState } from '../gen/common_pb'
import { gameClient } from '../lib/api-client'

/**
 * ゲームアクション実行のカスタムフック
 */
export const useGameActions = (
  gameId: string,
  currentPlayerId: string,
  onGameStateUpdate: (state: GameState) => void,
) => {
  const [message, setMessage] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  /**
   * カードをプレイする
   */
  const playCard = useCallback(
    async (cardId: string, targetId?: string) => {
      setIsLoading(true)
      try {
        const response = await gameClient.playCard({
          gameId,
          playerId: currentPlayerId,
          cardId,
          targetId,
        })

        if (response.success && response.gameState) {
          onGameStateUpdate(response.gameState)
          setMessage(response.message)
          return true
        } else {
          setMessage(response.message || 'カードのプレイに失敗しました')
          return false
        }
      } catch (err) {
        setMessage(
          err instanceof Error ? err.message : 'カードのプレイに失敗しました',
        )
        return false
      } finally {
        setIsLoading(false)
      }
    },
    [gameId, currentPlayerId, onGameStateUpdate],
  )

  /**
   * 攻撃を実行する
   */
  const executeAttack = useCallback(
    async (attackerId: string, targetId?: string) => {
      setIsLoading(true)
      try {
        const response = await gameClient.executeAttack({
          gameId,
          playerId: currentPlayerId,
          attackerId,
          targetId,
        })

        if (response.success && response.gameState) {
          onGameStateUpdate(response.gameState)
          setMessage(response.message)
          return true
        } else {
          setMessage(response.message || '攻撃に失敗しました')
          return false
        }
      } catch (err) {
        setMessage(err instanceof Error ? err.message : '攻撃に失敗しました')
        return false
      } finally {
        setIsLoading(false)
      }
    },
    [gameId, currentPlayerId, onGameStateUpdate],
  )

  /**
   * ターンを終了する
   */
  const endTurn = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await gameClient.endTurn({
        gameId,
        playerId: currentPlayerId,
      })

      if (response.success && response.gameState) {
        onGameStateUpdate(response.gameState)
        setMessage(response.message)
        return true
      }
      return false
    } catch (err) {
      setMessage(
        err instanceof Error ? err.message : 'ターン終了に失敗しました',
      )
      return false
    } finally {
      setIsLoading(false)
    }
  }, [gameId, currentPlayerId, onGameStateUpdate])

  /**
   * メッセージを自動で消す
   */
  const clearMessage = useCallback(() => {
    setMessage(null)
  }, [])

  return {
    playCard,
    executeAttack,
    endTurn,
    message,
    isLoading,
    clearMessage,
  }
}

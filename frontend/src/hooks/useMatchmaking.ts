import { useCallback, useState } from 'react'
import type { GameState } from '../gen/common_pb'
import { MatchmakingStatus } from '../gen/game_pb'
import { gameClient } from '../lib/api-client'

/**
 * プレイヤーIDを生成
 */
const generatePlayerId = (): string => {
  const timestamp = Date.now()
  const randomId = Math.random().toString(36).substring(2, 9)
  return `player-${timestamp}-${randomId}`
}

/**
 * マッチング処理のカスタムフック
 */
export const useMatchmaking = () => {
  const [isMatchmaking, setIsMatchmaking] = useState(false)
  const [matchmakingStatus, setMatchmakingStatus] = useState<string>('')
  const [error, setError] = useState<string | null>(null)
  const [abortController, setAbortController] =
    useState<AbortController | null>(null)

  /**
   * マッチングに参加
   */
  const joinMatchmaking = useCallback(
    async (
      playerName: string,
      onMatched: (gameState: GameState, playerId: string) => void,
      deckId?: string,
    ) => {
      if (!playerName.trim()) {
        setError('名前を入力してください')
        return
      }

      setIsMatchmaking(true)
      setError(null)
      setMatchmakingStatus('マッチング開始...')

      const playerId = generatePlayerId()
      const controller = new AbortController()
      setAbortController(controller)

      console.log('マッチング参加:', { playerId, playerName, deckId })

      try {
        const stream = gameClient.joinMatchmaking(
          {
            playerId,
            playerName,
            deckId: deckId || undefined,
          },
          { signal: controller.signal },
        )

        for await (const response of stream) {
          console.log('マッチングレスポンス:', response)

          switch (response.status) {
            case MatchmakingStatus.WAITING:
              setMatchmakingStatus(
                response.message || 'マッチング相手を探しています...',
              )
              break

            case MatchmakingStatus.MATCHED:
              setMatchmakingStatus(response.message || 'マッチング成功！')
              break

            case MatchmakingStatus.GAME_STARTED:
              setMatchmakingStatus('マリガン中...')
              if (response.gameState) {
                onMatched(response.gameState, playerId)
                setIsMatchmaking(false)
                setAbortController(null)
              }
              return

            default:
              setError(response.message || 'マッチングエラー')
              setIsMatchmaking(false)
              setAbortController(null)
              return
          }
        }
      } catch (err) {
        // キャンセルによるエラーかチェック
        const error = err as { name?: string; code?: string; message?: string }
        const isCanceled =
          error.name === 'AbortError' ||
          error.code === 'canceled' ||
          error.message?.includes('aborted') ||
          error.message?.includes('canceled')

        if (!isCanceled) {
          console.error('マッチングエラー:', err)
          setError(
            err instanceof Error ? err.message : 'マッチングに失敗しました',
          )
        } else {
          console.log('ユーザーによりマッチングがキャンセルされました')
        }
        setIsMatchmaking(false)
        setAbortController(null)
      }
    },
    [],
  )

  /**
   * マッチングをキャンセル
   */
  const cancelMatchmaking = useCallback(() => {
    if (abortController) {
      abortController.abort()
    }
    setIsMatchmaking(false)
    setMatchmakingStatus('')
    setError(null)
    setAbortController(null)
  }, [abortController])

  return {
    isMatchmaking,
    matchmakingStatus,
    error,
    joinMatchmaking,
    cancelMatchmaking,
  }
}

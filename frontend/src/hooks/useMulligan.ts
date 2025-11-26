import { useCallback, useState } from 'react'
import type { GameState } from '../gen/common_pb'
import { gameClient } from '../lib/api-client'

/**
 * マリガン処理のカスタムフック
 */
export const useMulligan = () => {
  const [isWaitingForOpponent, setIsWaitingForOpponent] = useState(false)
  const [error, setError] = useState<string | null>(null)

  /**
   * マリガン完了イベントを購読
   */
  const subscribeToMulliganCompletion = useCallback(
    async (
      gameId: string,
      playerId: string,
      onComplete: (gameState: GameState) => void,
    ) => {
      try {
        const stream = gameClient.streamGameEvents({
          gameId,
          playerId,
        })

        for await (const response of stream) {
          console.log('ゲームイベント:', response)

          // ターン開始イベントでゲームを開始
          if (
            response.event?.eventType === 'turn_started' &&
            response.gameState
          ) {
            console.log('ゲーム開始')
            onComplete(response.gameState)
            return
          }
        }
      } catch (err) {
        console.error('ストリームエラー:', err)
        setError('ゲーム開始に失敗しました')
      }
    },
    [],
  )

  /**
   * マリガンを実行
   */
  const performMulligan = useCallback(
    async (
      gameId: string,
      playerId: string,
      selectedCardIds: string[],
      onComplete: (gameState: GameState) => void,
    ) => {
      try {
        const response = await gameClient.performMulligan({
          gameId,
          playerId,
          cardIds: selectedCardIds,
        })

        console.log('マリガンレスポンス:', response)

        if (response.success && response.gameState) {
          const bothDone =
            response.gameState.player1MulliganDone &&
            response.gameState.player2MulliganDone

          if (bothDone) {
            console.log(
              '両プレイヤーのマリガンが完了しました。ゲームを開始します。',
            )

            // サーバー側でStartTurnが実行されるのを待つ
            await new Promise((resolve) => setTimeout(resolve, 300))

            try {
              const latestStateResponse = await gameClient.getGameState({
                gameId: response.gameState.gameId,
                playerId,
              })

              if (latestStateResponse.gameState) {
                onComplete(latestStateResponse.gameState)
              } else {
                onComplete(response.gameState)
              }
            } catch (err) {
              console.error('ゲーム状態取得エラー:', err)
              onComplete(response.gameState)
            }
          } else {
            // 相手が完了していない場合、待機状態にして購読開始
            setIsWaitingForOpponent(true)
            await subscribeToMulliganCompletion(gameId, playerId, onComplete)
          }
        } else {
          setError(response.message || 'マリガンに失敗しました')
        }
      } catch (err) {
        console.error('マリガンエラー:', err)
        setError(err instanceof Error ? err.message : 'マリガンに失敗しました')
      }
    },
    [subscribeToMulliganCompletion],
  )

  return {
    performMulligan,
    isWaitingForOpponent,
    error,
  }
}

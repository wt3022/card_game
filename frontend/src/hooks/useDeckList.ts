import { useCallback, useEffect, useState } from 'react'
import type { Deck } from '../gen/common_pb'
import { publicCardManagementClient } from '../lib/api-client'

export function useDeckList(userId?: string) {
  const [decks, setDecks] = useState<Deck[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchDecks = useCallback(async (fetchUserId?: string) => {
    setIsLoading(true)
    setError(null)

    try {
      console.log('デッキ取得リクエスト - userId:', fetchUserId)
      const response = await publicCardManagementClient.listDecks({
        userId: fetchUserId,
      })
      console.log(
        'デッキ取得成功 - 件数:',
        response.decks.length,
        'decks:',
        response.decks,
      )
      setDecks(response.decks)
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'デッキの取得に失敗しました'
      setError(errorMessage)
      console.error('デッキ取得エラー:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    console.log('useDeckList useEffect - userId:', userId)
    fetchDecks(userId)
  }, [fetchDecks, userId])

  return {
    decks,
    isLoading,
    error,
    refetch: fetchDecks,
  }
}

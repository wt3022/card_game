import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { useCallback, useEffect, useState } from 'react'
import { CardManagementService } from '../gen/card_management_connect'
import type { Deck } from '../gen/common_pb'

const transport = createConnectTransport({
  baseUrl: 'http://localhost:8080',
})

const client = createPromiseClient(CardManagementService, transport)

export function useDeckList() {
  const [decks, setDecks] = useState<Deck[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchDecks = useCallback(async (userId?: string) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await client.listDecks({
        userId: userId,
      })
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
    fetchDecks()
  }, [fetchDecks])

  return {
    decks,
    isLoading,
    error,
    refetch: fetchDecks,
  }
}

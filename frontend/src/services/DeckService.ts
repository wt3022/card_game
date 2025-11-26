/**
 * サービス層: デッキ管理サービス
 * APIクライアントをドメインモデルでラップ
 */

import type { Deck } from '../domain/models/Deck'
import { cardManagementClient } from '../lib/api-client'
import { DeckMapper } from '../mappers'

/**
 * デッキ管理サービス
 */
// biome-ignore lint/complexity/noStaticOnlyClass: Serviceクラスは名前空間として使用
export class DeckService {
  /**
   * ユーザーのデッキ一覧を取得
   */
  static async listDecks(userId: string): Promise<Deck[]> {
    const response = await cardManagementClient.listDecks({ userId })
    return DeckMapper.toDomainArray(response.decks)
  }

  /**
   * デッキIDで取得
   */
  static async getDeck(deckId: string): Promise<Deck> {
    const response = await cardManagementClient.getDeck({ id: deckId })
    if (!response.deck) {
      throw new Error(`Deck not found: ${deckId}`)
    }
    return DeckMapper.toDomain(response.deck)
  }

  /**
   * デッキを作成
   */
  static async createDeck(
    deck: Omit<Deck, 'id' | 'createdAt' | 'updatedAt'>,
  ): Promise<Deck> {
    const response = await cardManagementClient.createDeck({
      name: deck.name,
      description: deck.description,
      cardIds: deck.cardIds,
    })
    if (!response.deck) {
      throw new Error('Failed to create deck')
    }
    return DeckMapper.toDomain(response.deck)
  }

  /**
   * デッキを更新
   */
  static async updateDeck(deck: Deck): Promise<Deck> {
    const response = await cardManagementClient.updateDeck({
      id: deck.id,
      name: deck.name,
      description: deck.description,
      cardIds: deck.cardIds,
    })
    if (!response.deck) {
      throw new Error(`Failed to update deck: ${deck.id}`)
    }
    return DeckMapper.toDomain(response.deck)
  }

  /**
   * デッキを削除
   */
  static async deleteDeck(deckId: string): Promise<void> {
    await cardManagementClient.deleteDeck({ id: deckId })
  }
}

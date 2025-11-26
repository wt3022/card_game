/**
 * サービス層: カード管理サービス
 * APIクライアントをドメインモデルでラップ
 */

import type { Card } from '../domain/models/Card'
import { cardManagementClient } from '../lib/api-client'
import { CardMapper } from '../mappers'

/**
 * カード管理サービス
 */
// biome-ignore lint/complexity/noStaticOnlyClass: Serviceクラスは名前空間として使用
export class CardService {
  /**
   * すべてのカードを取得
   */
  static async listCards(): Promise<Card[]> {
    const response = await cardManagementClient.listCards({})
    return response.cards.map((c) => CardMapper.toDomain(c))
  }

  /**
   * カードIDで取得
   */
  static async getCard(cardId: string): Promise<Card> {
    const response = await cardManagementClient.getCard({ id: cardId })
    if (!response.card) {
      throw new Error(`Card not found: ${cardId}`)
    }
    return CardMapper.toDomain(response.card)
  }

  /**
   * カードを作成
   */
  static async createCard(card: Omit<Card, 'id'>): Promise<Card> {
    const protoCard = CardMapper.toProto({ ...card, id: '' })
    const response = await cardManagementClient.createCard(protoCard)
    if (!response.card) {
      throw new Error('Failed to create card')
    }
    return CardMapper.toDomain(response.card)
  }

  /**
   * カードを更新
   */
  static async updateCard(card: Card): Promise<Card> {
    const protoCard = CardMapper.toProto(card)
    const response = await cardManagementClient.updateCard(protoCard)
    if (!response.card) {
      throw new Error(`Failed to update card: ${card.id}`)
    }
    return CardMapper.toDomain(response.card)
  }

  /**
   * カードを削除
   */
  static async deleteCard(cardId: string): Promise<void> {
    await cardManagementClient.deleteCard({ id: cardId })
  }
}

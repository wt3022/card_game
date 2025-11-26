/**
 * ドメイン層: デッキモデル
 */

export interface Deck {
  id: string
  name: string
  description: string
  cardIds: string[]
  userId: string
  createdAt: Date
  updatedAt: Date
}

export const MIN_DECK_SIZE = 30
export const MAX_DECK_SIZE = 30
export const MAX_CARD_COPIES = 3

export function validateDeck(deck: Partial<Deck>): string[] {
  const errors: string[] = []

  if (!deck.name || deck.name.trim() === '') {
    errors.push('デッキ名は必須です')
  }

  if (!deck.cardIds || deck.cardIds.length === 0) {
    errors.push('デッキには最低1枚のカードが必要です')
  }

  if (deck.cardIds && deck.cardIds.length < MIN_DECK_SIZE) {
    errors.push(`デッキには${MIN_DECK_SIZE}枚のカードが必要です`)
  }

  if (deck.cardIds && deck.cardIds.length > MAX_DECK_SIZE) {
    errors.push(`デッキには${MAX_DECK_SIZE}枚までしか入れられません`)
  }

  // 同じカードのコピー数チェック
  if (deck.cardIds) {
    const cardCounts = new Map<string, number>()
    for (const cardId of deck.cardIds) {
      cardCounts.set(cardId, (cardCounts.get(cardId) || 0) + 1)
    }

    for (const [cardId, count] of cardCounts.entries()) {
      if (count > MAX_CARD_COPIES) {
        errors.push(
          `カード ${cardId} は${MAX_CARD_COPIES}枚までしか入れられません`,
        )
      }
    }
  }

  return errors
}

export function isValidDeck(deck: Partial<Deck>): boolean {
  return validateDeck(deck).length === 0
}

export interface Deck {
  id: string
  name: string
  description: string
  cardIds: string[]
  createdAt?: string
  updatedAt?: string
}

export interface DeckWithCards extends Deck {
  cards: Array<{
    id: string
    name: string
    cost: number
    count: number
  }>
}

/**
 * ドメイン層: ゲームモデル
 */

import type { Card } from './Card'

export type GamePhase =
  | 'TurnStart'
  | 'Draw'
  | 'ResourceGain'
  | 'Main'
  | 'TurnEnd'

export interface Unit {
  cardId: string
  instanceId: string
  name: string
  cost: number
  attack: number
  defense: number
  currentDefense: number
  traits: string[]
  effect: string
  attacksRemaining: number
  summonedThisTurn: boolean
  ownerId: string
}

export interface Player {
  id: string
  name: string
  hp: number
  maxHp: number
  currentTurnMana: number
  currentRecoveryMana: number
  handCount: number
  deckCount: number
  graveyardCount: number
  field: Unit[]
  hand?: Card[]
  timeRemainingSeconds: number
  isConnected: boolean
  lastActivityAt: Date
}

export interface GameState {
  gameId: string
  player1: Player
  player2: Player
  currentPlayerId: string
  currentTurn: number
  currentPhase: GamePhase
  isGameOver: boolean
  winnerId?: string
  isDraw: boolean
  player1MulliganDone: boolean
  player2MulliganDone: boolean
}

export interface GameEvent {
  gameId: string
  eventType: string
  message: string
  playerId?: string
  timestamp: Date
  gameState?: GameState
}

// ゲームロジックヘルパー
export function isMyTurn(gameState: GameState, playerId: string): boolean {
  return gameState.currentPlayerId === playerId
}

export function getMyPlayer(
  gameState: GameState,
  playerId: string,
): Player | null {
  if (gameState.player1.id === playerId) return gameState.player1
  if (gameState.player2.id === playerId) return gameState.player2
  return null
}

export function getOpponentPlayer(
  gameState: GameState,
  playerId: string,
): Player | null {
  if (gameState.player1.id === playerId) return gameState.player2
  if (gameState.player2.id === playerId) return gameState.player1
  return null
}

export function canPlayCard(card: Card, player: Player): boolean {
  return card.cost <= player.currentTurnMana
}

export function canAttack(unit: Unit): boolean {
  return unit.attacksRemaining > 0 && !unit.summonedThisTurn
}

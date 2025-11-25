import type { GameState, Player } from '../gen/common_pb'

/**
 * 現在のプレイヤーを取得
 */
export const getCurrentPlayer = (
  gameState: GameState,
  currentPlayerId: string,
): Player | undefined => {
  return gameState.player1?.id === currentPlayerId
    ? gameState.player1
    : gameState.player2
}

/**
 * 相手プレイヤーを取得
 */
export const getOpponent = (
  gameState: GameState,
  currentPlayerId: string,
): Player | undefined => {
  return gameState.player1?.id === currentPlayerId
    ? gameState.player2
    : gameState.player1
}

/**
 * 現在のプレイヤーのターンかどうかを判定
 */
export const isCurrentPlayerTurn = (
  gameState: GameState,
  currentPlayerId: string,
): boolean => {
  return gameState.currentPlayerId === currentPlayerId
}

/**
 * ユニットが自分のものかどうかを判定
 */
export const isMyUnit = (
  player: Player | undefined,
  unitId: string,
): boolean => {
  return player?.field.some((u) => u.instanceId === unitId) ?? false
}

/**
 * マリガン用の手札を取得
 */
export const getMulliganHand = (gameState: GameState, playerId: string) => {
  if (gameState.player1?.id === playerId) {
    return gameState.player1.hand || []
  } else if (gameState.player2?.id === playerId) {
    return gameState.player2.hand || []
  }
  return []
}

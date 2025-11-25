/**
 * コンポーネント共通の型定義
 */

import type { GameState, Player } from '../gen/common_pb'

/**
 * ゲームボードのプロパティ
 */
export interface GameBoardProps {
  gameState: GameState
  currentPlayerId: string
  onGameStateUpdate: (gameState: GameState) => void
}

/**
 * ゲームセットアップのプロパティ
 */
export interface GameSetupProps {
  onGameStart: (gameState: GameState, playerId: string) => void
}

/**
 * プレイヤー情報のプロパティ
 */
export interface PlayerInfoProps {
  player: Player
  isCurrentPlayer: boolean
  onClick: () => void
  isAttackTarget?: boolean
}

/**
 * マリガンモーダルのプロパティ
 */
export interface MulliganModalProps {
  hand: Array<{ id: string; name: string; cost: number }>
  onSubmit: (selectedCardIds: string[]) => void
  onSkip: () => void
  isWaitingForOpponent: boolean
}

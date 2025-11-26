/**
 * サービス層: ゲームサービス
 * APIクライアントをドメインモデルでラップ
 */

import type { GameState } from '../domain/models/Game'
import { gameClient } from '../lib/api-client'
import { GameMapper } from '../mappers'

/**
 * ゲームサービス
 */
// biome-ignore lint/complexity/noStaticOnlyClass: Serviceクラスは名前空間として使用
export class GameService {
  /**
   * マッチメイキングに参加 (StreamingRPC)
   */
  static joinMatchmaking(
    playerId: string,
    playerName: string,
    deckId: string | undefined,
    onResponse: (status: number, gameId?: string) => void,
    onError: (error: Error) => void,
    signal?: AbortSignal,
  ) {
    const stream = gameClient.joinMatchmaking(
      { playerId, playerName, deckId },
      { signal },
    )

    ;(async () => {
      try {
        for await (const response of stream) {
          onResponse(response.status, response.gameId)
        }
      } catch (error) {
        if (error instanceof Error) {
          onError(error)
        } else {
          onError(new Error('Unknown error occurred'))
        }
      }
    })()
  }

  /**
   * ゲーム状態を取得
   */
  static async getGameState(
    gameId: string,
    playerId: string,
  ): Promise<GameState> {
    const response = await gameClient.getGameState({ gameId, playerId })
    if (!response.gameState) {
      throw new Error(`Game state not found: ${gameId}`)
    }
    return GameMapper.gameStateToDomain(response.gameState)
  }

  /**
   * ゲームイベントをストリーミング購読
   */
  static streamGameEvents(
    gameId: string,
    playerId: string,
    onEvent: (gameState: GameState) => void,
    onError: (error: Error) => void,
    signal?: AbortSignal,
  ) {
    const stream = gameClient.streamGameEvents({ gameId, playerId }, { signal })

    ;(async () => {
      try {
        for await (const response of stream) {
          if (response.gameState) {
            const gameState = GameMapper.gameStateToDomain(response.gameState)
            onEvent(gameState)
          }
        }
      } catch (error) {
        if (error instanceof Error) {
          onError(error)
        } else {
          onError(new Error('Unknown error occurred'))
        }
      }
    })()
  }

  /**
   * カードをプレイ
   */
  static async playCard(
    gameId: string,
    playerId: string,
    cardId: string,
    targetId?: string,
  ): Promise<GameState> {
    const response = await gameClient.playCard({
      gameId,
      playerId,
      cardId,
      targetId,
    })
    if (!response.gameState) {
      throw new Error('Failed to play card')
    }
    return GameMapper.gameStateToDomain(response.gameState)
  }

  /**
   * ユニットで攻撃
   */
  static async executeAttack(
    gameId: string,
    playerId: string,
    attackerId: string,
    targetId: string,
  ): Promise<GameState> {
    const response = await gameClient.executeAttack({
      gameId,
      playerId,
      attackerId,
      targetId,
    })
    if (!response.gameState) {
      throw new Error('Failed to attack')
    }
    return GameMapper.gameStateToDomain(response.gameState)
  }

  /**
   * ターンをエンド
   */
  static async endTurn(gameId: string, playerId: string): Promise<GameState> {
    const response = await gameClient.endTurn({ gameId, playerId })
    if (!response.gameState) {
      throw new Error('Failed to end turn')
    }
    return GameMapper.gameStateToDomain(response.gameState)
  }

  /**
   * マリガンを実行
   */
  static async performMulligan(
    gameId: string,
    playerId: string,
    cardIds: string[],
  ): Promise<GameState> {
    const response = await gameClient.performMulligan({
      gameId,
      playerId,
      cardIds,
    })
    if (!response.gameState) {
      throw new Error('Failed to perform mulligan')
    }
    return GameMapper.gameStateToDomain(response.gameState)
  }
}

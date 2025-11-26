/**
 * マッパー層: Proto ↔ Domain 変換
 * ゲームステートモデルの変換ロジック
 */

import { Timestamp } from '@bufbuild/protobuf'
import type {
  GameEvent,
  GamePhase,
  GameState,
  Player,
  Unit,
} from '../domain/models/Game'
import {
  type GameEvent as ProtoGameEvent,
  GamePhase as ProtoGamePhase,
  type GameState as ProtoGameState,
  type Player as ProtoPlayer,
  type Unit as ProtoUnit,
} from '../gen/common_pb'

import { CardMapper } from './cardMapper'

// biome-ignore lint/complexity/noStaticOnlyClass: Mapperクラスは名前空間として使用
export class GameMapper {
  // GameState: Proto → Domain
  static gameStateToDomain(proto: ProtoGameState): GameState {
    return {
      gameId: proto.gameId,
      player1: proto.player1
        ? GameMapper.playerToDomain(proto.player1)
        : GameMapper.createEmptyPlayer('1'),
      player2: proto.player2
        ? GameMapper.playerToDomain(proto.player2)
        : GameMapper.createEmptyPlayer('2'),
      currentPlayerId: proto.currentPlayerId,
      currentTurn: proto.currentTurn,
      currentPhase: GameMapper.gamePhaseToDomain(proto.currentPhase),
      isGameOver: proto.isGameOver,
      winnerId: proto.winnerId,
      isDraw: proto.isDraw,
      player1MulliganDone: proto.player1MulliganDone,
      player2MulliganDone: proto.player2MulliganDone,
    }
  }

  // GameState: Domain → Proto
  static gameStateToProto(domain: GameState): Partial<ProtoGameState> {
    return {
      gameId: domain.gameId,
      player1: domain.player1
        ? (GameMapper.playerToProto(domain.player1) as ProtoPlayer)
        : undefined,
      player2: domain.player2
        ? (GameMapper.playerToProto(domain.player2) as ProtoPlayer)
        : undefined,
      currentPlayerId: domain.currentPlayerId,
      currentTurn: domain.currentTurn,
      currentPhase: GameMapper.gamePhaseToProto(domain.currentPhase),
      isGameOver: domain.isGameOver,
      winnerId: domain.winnerId,
      isDraw: domain.isDraw,
      player1MulliganDone: domain.player1MulliganDone,
      player2MulliganDone: domain.player2MulliganDone,
    }
  }

  // Player: Proto → Domain
  static playerToDomain(proto: ProtoPlayer): Player {
    return {
      id: proto.id,
      name: proto.name,
      hp: proto.hp,
      maxHp: proto.maxHp,
      currentTurnMana: proto.currentTurnMana,
      currentRecoveryMana: proto.currentRecoveryMana,
      hand: proto.hand.map((c) =>
        CardMapper.toDomain(c as ProtoPlayer['hand'][number]),
      ),
      field: proto.field.map((u: ProtoUnit) => GameMapper.unitToDomain(u)),
      handCount: proto.handCount,
      deckCount: proto.deckCount,
      graveyardCount: proto.graveyardCount,
      timeRemainingSeconds: proto.timeRemainingSeconds,
      isConnected: proto.isConnected,
      lastActivityAt: proto.lastActivityAt?.toDate() || new Date(),
    }
  }

  // Player: Domain → Proto
  static playerToProto(domain: Player): Partial<ProtoPlayer> {
    return {
      id: domain.id,
      name: domain.name,
      hp: domain.hp,
      maxHp: domain.maxHp,
      currentTurnMana: domain.currentTurnMana,
      currentRecoveryMana: domain.currentRecoveryMana,
      hand: (domain.hand || []).map((c) =>
        CardMapper.toProto(c),
      ) as ProtoPlayer['hand'],
      field: domain.field.map((u) =>
        GameMapper.unitToProto(u),
      ) as ProtoPlayer['field'],
      handCount: domain.handCount,
      deckCount: domain.deckCount,
      graveyardCount: domain.graveyardCount,
      timeRemainingSeconds: domain.timeRemainingSeconds,
      isConnected: domain.isConnected,
    }
  }

  // Unit: Proto → Domain
  static unitToDomain(proto: ProtoUnit): Unit {
    return {
      cardId: proto.cardId,
      instanceId: proto.instanceId,
      name: proto.name,
      cost: proto.cost,
      attack: proto.attack,
      defense: proto.defense,
      currentDefense: proto.currentDefense,
      traits: proto.traits.map((t) =>
        CardMapper.traitToDomain(t as ProtoUnit['traits'][number]),
      ),
      effect: proto.effect,
      attacksRemaining: proto.attacksRemaining,
      summonedThisTurn: proto.summonedThisTurn,
      ownerId: proto.ownerId,
    }
  }

  // Unit: Domain → Proto
  static unitToProto(domain: Unit): Partial<ProtoUnit> {
    return {
      cardId: domain.cardId,
      instanceId: domain.instanceId,
      name: domain.name,
      cost: domain.cost,
      attack: domain.attack,
      defense: domain.defense,
      currentDefense: domain.currentDefense,
      traits: domain.traits.map((t) =>
        CardMapper.traitToProto(
          t as Parameters<typeof CardMapper.traitToProto>[0],
        ),
      ),
      effect: domain.effect,
      attacksRemaining: domain.attacksRemaining,
      summonedThisTurn: domain.summonedThisTurn,
      ownerId: domain.ownerId,
    }
  }

  // GameEvent: Proto → Domain
  static gameEventToDomain(proto: ProtoGameEvent): GameEvent {
    return {
      gameId: proto.gameId,
      eventType: proto.eventType,
      message: proto.details,
      playerId: proto.playerId,
      timestamp: proto.timestamp?.toDate() || new Date(),
    }
  }

  // GameEvent: Domain → Proto
  static gameEventToProto(domain: GameEvent): Partial<ProtoGameEvent> {
    return {
      gameId: domain.gameId,
      eventType: domain.eventType,
      details: domain.message,
      playerId: domain.playerId,
      timestamp: Timestamp.fromDate(domain.timestamp),
    }
  }

  // GamePhase: Proto → Domain
  private static gamePhaseToDomain(proto: ProtoGamePhase): GamePhase {
    switch (proto) {
      case ProtoGamePhase.TURN_START:
        return 'TurnStart'
      case ProtoGamePhase.DRAW:
        return 'Draw'
      case ProtoGamePhase.RESOURCE_GAIN:
        return 'ResourceGain'
      case ProtoGamePhase.MAIN:
        return 'Main'
      case ProtoGamePhase.TURN_END:
        return 'TurnEnd'
      default:
        return 'TurnStart'
    }
  }

  // GamePhase: Domain → Proto
  private static gamePhaseToProto(domain: GamePhase): ProtoGamePhase {
    switch (domain) {
      case 'TurnStart':
        return ProtoGamePhase.TURN_START
      case 'Draw':
        return ProtoGamePhase.DRAW
      case 'ResourceGain':
        return ProtoGamePhase.RESOURCE_GAIN
      case 'Main':
        return ProtoGamePhase.MAIN
      case 'TurnEnd':
        return ProtoGamePhase.TURN_END
    }
  }

  // ヘルパーメソッド: 空のプレイヤーを作成
  private static createEmptyPlayer(id: string): Player {
    return {
      id,
      name: 'Unknown',
      hp: 0,
      maxHp: 0,
      currentTurnMana: 0,
      currentRecoveryMana: 0,
      hand: [],
      field: [],
      handCount: 0,
      deckCount: 0,
      graveyardCount: 0,
      timeRemainingSeconds: 0,
      isConnected: false,
      lastActivityAt: new Date(),
    }
  }
}

/**
 * マッパー層: Proto ↔ Domain 変換
 * デッキモデルの変換ロジック
 */

import type { Deck } from '../../domain/models/Deck'
import type { Deck as ProtoDeck } from '../../gen/card_management_pb'

// Proto → Domain
export function deckToDomain(proto: ProtoDeck): Deck {
  return {
    id: proto.id,
    name: proto.name,
    description: proto.description,
    cardIds: [...proto.cardIds],
    userId: proto.userId,
    createdAt: proto.createdAt
      ? new Date(Number(proto.createdAt) * 1000)
      : undefined,
    updatedAt: proto.updatedAt
      ? new Date(Number(proto.updatedAt) * 1000)
      : undefined,
  }
}

// Domain → Proto
export function deckToProto(domain: Deck): Partial<ProtoDeck> {
  return {
    id: domain.id,
    name: domain.name,
    description: domain.description,
    cardIds: [...domain.cardIds],
    userId: domain.userId,
    createdAt: domain.createdAt
      ? BigInt(Math.floor(domain.createdAt.getTime() / 1000))
      : undefined,
    updatedAt: domain.updatedAt
      ? BigInt(Math.floor(domain.updatedAt.getTime() / 1000))
      : undefined,
  }
}

// Proto配列 → Domain配列
export function deckArrayToDomain(protos: ProtoDeck[]): Deck[] {
  return protos.map((p) => deckToDomain(p))
}

// Domain配列 → Proto配列
export function deckArrayToProto(domains: Deck[]): Partial<ProtoDeck>[] {
  return domains.map((d) => deckToProto(d))
}

// 後方互換性のため、従来のクラスベースのインターフェイスも提供
export const DeckMapper = {
  toDomain: deckToDomain,
  toProto: deckToProto,
  toDomainArray: deckArrayToDomain,
  toProtoArray: deckArrayToProto,
}

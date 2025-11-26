/**
 * マッパー層: Proto ↔ Domain 変換
 * デッキモデルの変換ロジック
 */

import { Timestamp } from '@bufbuild/protobuf'
import type { Deck } from '../domain/models/Deck'
import type { Deck as ProtoDeck } from '../gen/common_pb'

// Proto → Domain
export function deckToDomain(proto: ProtoDeck): Deck {
  return {
    id: proto.id,
    name: proto.name,
    description: proto.description,
    cardIds: [...proto.cardIds],
    userId: proto.userId,
    createdAt: proto.createdAt ? proto.createdAt.toDate() : new Date(),
    updatedAt: proto.updatedAt ? proto.updatedAt.toDate() : new Date(),
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
    createdAt: Timestamp.fromDate(domain.createdAt),
    updatedAt: Timestamp.fromDate(domain.updatedAt),
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

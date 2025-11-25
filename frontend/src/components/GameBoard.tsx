import { useEffect, useState } from 'react'
import { MESSAGE_DISPLAY_DURATION, TRAIT_LABELS } from '../constants/game'
import { useGameActions } from '../hooks/useGameActions'
import { useGameState } from '../hooks/useGameState'
import type { GameBoardProps } from '../types/components'
import {
  isCurrentPlayerTurn as checkIsCurrentPlayerTurn,
  getCurrentPlayer,
  getOpponent,
  isMyUnit,
} from '../utils/gameHelpers'
import PlayerInfo from './PlayerInfo'
import UnitCard from './UnitCard'
import './GameBoard.css'

export default function GameBoard({
  gameState: initialGameState,
  currentPlayerId,
  onGameStateUpdate,
}: GameBoardProps) {
  const [selectedCardId, setSelectedCardId] = useState<string | null>(null)
  const [selectedUnitId, setSelectedUnitId] = useState<string | null>(null)

  // ゲーム状態の購読と管理
  const { gameState: liveGameState } = useGameState(
    initialGameState.gameId,
    currentPlayerId,
  )

  // 最新のゲーム状態を使用（リアルタイム更新があればそれを、なければ初期状態を使用）
  const gameState = liveGameState || initialGameState

  // ゲームアクション
  const { playCard, executeAttack, endTurn, message, clearMessage } =
    useGameActions(gameState.gameId, currentPlayerId, onGameStateUpdate)

  // 状態の取得
  const isCurrentPlayerTurn = checkIsCurrentPlayerTurn(
    gameState,
    currentPlayerId,
  )
  const currentPlayer = getCurrentPlayer(gameState, currentPlayerId)
  const opponent = getOpponent(gameState, currentPlayerId)

  // メッセージを自動で消す
  useEffect(() => {
    if (message) {
      const timer = setTimeout(clearMessage, MESSAGE_DISPLAY_DURATION)
      return () => clearTimeout(timer)
    }
  }, [message, clearMessage])

  // ゲーム状態が更新されたら親に通知
  useEffect(() => {
    if (liveGameState) {
      onGameStateUpdate(liveGameState)
    }
  }, [liveGameState, onGameStateUpdate])

  /**
   * カードプレイのハンドラー
   */
  const handlePlayCard = async (cardId: string, targetId?: string) => {
    const success = await playCard(cardId, targetId)
    if (success) {
      setSelectedCardId(null)
    }
  }

  /**
   * 攻撃のハンドラー
   */
  const handleAttack = async (attackerId: string, targetId?: string) => {
    const success = await executeAttack(attackerId, targetId)
    if (success) {
      setSelectedUnitId(null)
    }
  }

  /**
   * ターン終了のハンドラー
   */
  const handleEndTurn = async () => {
    await endTurn()
    setSelectedCardId(null)
    setSelectedUnitId(null)
  }

  /**
   * ユニットクリックのハンドラー
   */
  const handleUnitClick = (unitId: string) => {
    if (!isCurrentPlayerTurn) return

    // 自分のユニットをクリックした場合
    if (isMyUnit(currentPlayer, unitId)) {
      setSelectedUnitId(unitId === selectedUnitId ? null : unitId)
      setSelectedCardId(null)
    }
    // 攻撃対象として相手のユニットをクリックした場合
    else if (selectedUnitId) {
      handleAttack(selectedUnitId, unitId)
    }
  }

  /**
   * 相手プレイヤーへの直接攻撃
   */
  const handleOpponentPlayerClick = () => {
    if (!isCurrentPlayerTurn || !selectedUnitId) return
    handleAttack(selectedUnitId)
  }

  // ゲーム終了時の表示
  if (gameState.isGameOver) {
    return (
      <div className="game-over">
        <h2>
          {gameState.isDraw
            ? '引き分け'
            : `${gameState.winnerId === currentPlayerId ? '勝利' : '敗北'}！`}
        </h2>
        <button type="button" onClick={() => window.location.reload()}>
          新しいゲーム
        </button>
      </div>
    )
  }

  return (
    <div className="game-board">
      {message && <div className="message-banner">{message}</div>}

      {/* 相手の情報 */}
      <div className="opponent-area">
        {opponent && (
          <PlayerInfo
            player={opponent}
            isCurrentPlayer={!isCurrentPlayerTurn}
            onClick={handleOpponentPlayerClick}
            isAttackTarget={!!selectedUnitId}
          />
        )}
        <div className="field">
          {opponent?.field.map((unit) => (
            <UnitCard
              key={unit.instanceId}
              unit={unit}
              onClick={() => {
                if (selectedCardId) {
                  handlePlayCard(selectedCardId, unit.instanceId)
                  setSelectedCardId(null)
                } else {
                  handleUnitClick(unit.instanceId)
                }
              }}
              isSelected={false}
              isClickable={!!selectedUnitId || !!selectedCardId}
            />
          ))}
        </div>
      </div>

      {/* ゲーム情報 */}
      <div className="game-info">
        <div className="turn-info">
          <span>ターン {gameState.currentTurn}</span>
        </div>
        {isCurrentPlayerTurn && (
          <button
            type="button"
            className="end-turn-button"
            onClick={handleEndTurn}
          >
            ターン終了
          </button>
        )}
      </div>

      {/* 自分のフィールド */}
      <div className="player-area">
        <div className="field">
          {currentPlayer?.field.map((unit) => (
            <UnitCard
              key={unit.instanceId}
              unit={unit}
              onClick={() => {
                if (selectedCardId) {
                  handlePlayCard(selectedCardId, unit.instanceId)
                  setSelectedCardId(null)
                } else {
                  handleUnitClick(unit.instanceId)
                }
              }}
              isSelected={selectedUnitId === unit.instanceId}
              isClickable={isCurrentPlayerTurn}
            />
          ))}
        </div>
      </div>

      {/* 手札と自分の情報 */}
      <div className="hand-area">
        {currentPlayer && (
          <PlayerInfo
            player={currentPlayer}
            isCurrentPlayer={isCurrentPlayerTurn}
            onClick={() => {}}
            isAttackTarget={false}
          />
        )}
        <div className="hand-cards">
          {currentPlayer?.hand && currentPlayer.hand.length > 0 ? (
            currentPlayer.hand.map((card) => (
              <div
                key={card.id}
                className={`hand-card-detailed ${selectedCardId === card.id ? 'selected' : ''}`}
                role="button"
                tabIndex={0}
                onClick={() =>
                  setSelectedCardId(selectedCardId === card.id ? null : card.id)
                }
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    setSelectedCardId(
                      selectedCardId === card.id ? null : card.id,
                    )
                  }
                }}
                title={`${card.name} - コスト:${card.cost} ${card.effect}`}
              >
                <div className="card-cost">{card.cost}</div>
                <div className="card-name">{card.name}</div>
                {card.effect && (
                  <div className="card-effect">{card.effect}</div>
                )}
                {card.attack !== undefined && card.attack !== null && (
                  <div className="card-stats">
                    <span className="atk">{card.attack}</span>
                    <span className="def">{card.defense}</span>
                  </div>
                )}
                {card.traits && card.traits.length > 0 && (
                  <div className="card-traits">
                    {card.traits.map((trait) => (
                      <span key={trait} className="trait-badge-small">
                        {TRAIT_LABELS[trait] || trait}
                      </span>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="card-play-button"
                  onClick={(e) => {
                    e.stopPropagation()
                    handlePlayCard(card.id)
                  }}
                >
                  使用
                </button>
              </div>
            ))
          ) : (
            <div className="no-cards">手札がありません</div>
          )}
        </div>
      </div>
    </div>
  )
}

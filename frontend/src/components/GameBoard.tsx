import { useState, useEffect } from 'react'
import { gameClient } from '../lib/api-client'
import type { GameState } from '../gen/common_pb'
import { Trait } from '../gen/common_pb'
import PlayerInfo from './PlayerInfo'
import UnitCard from './UnitCard'
import './GameBoard.css'

interface GameBoardProps {
  gameState: GameState
  currentPlayerId: string
  onGameStateUpdate: (gameState: GameState) => void
}

export default function GameBoard({
  gameState,
  currentPlayerId,
  onGameStateUpdate,
}: GameBoardProps) {
  const [selectedCardId, setSelectedCardId] = useState<string | null>(null)
  const [selectedUnitId, setSelectedUnitId] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)

  // デバッグ: gameStateの内容を確認
  console.log('GameBoard - gameState:', gameState)
  console.log('GameBoard - currentPlayerId:', currentPlayerId)
  console.log('GameBoard - player1:', gameState.player1)
  console.log('GameBoard - player2:', gameState.player2)

  const isCurrentPlayerTurn = gameState.currentPlayerId === currentPlayerId
  const currentPlayer = gameState.player1?.id === currentPlayerId ? gameState.player1 : gameState.player2
  const opponent = gameState.player1?.id === currentPlayerId ? gameState.player2 : gameState.player1

  console.log('GameBoard - currentPlayer:', currentPlayer)
  console.log('GameBoard - opponent:', opponent)

  // 特性ラベル
  const traitLabels: Record<number, string> = {
    [Trait.RUSH]: '疾走',
    [Trait.CHARGE]: '突進',
    [Trait.WINDFURY]: '疾風',
    [Trait.PIERCE]: '貫通',
    [Trait.GUARDIAN]: '守護',
    [Trait.EFFECT_SHIELD]: '効果盾',
    [Trait.UNTARGETABLE]: '対象不可',
  }

  // メッセージを表示してから消す
  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => setMessage(null), 3000)
      return () => clearTimeout(timer)
    }
  }, [message])

  // リアルタイム更新: StreamGameEventsを購読（サーバーサイドストリーミング）
  useEffect(() => {
    let abortController: AbortController | null = null
    let isActive = true

    const subscribeToEvents = async () => {
      try {
        console.log('Subscribing to game events for gameId:', gameState.gameId)
        
        abortController = new AbortController()
        
        // サーバーサイドストリーミングでイベントを購読
        const stream = gameClient.streamGameEvents(
          {
            gameId: gameState.gameId,
            playerId: currentPlayerId,
          },
          { signal: abortController.signal }
        )

        // イベントを受信
        for await (const response of stream) {
          if (!isActive) break
          
          console.log('Received game event:', response)
          
          if (response.gameState) {
            onGameStateUpdate(response.gameState)
            
            // イベント詳細があれば表示
            if (response.event?.details) {
              setMessage(response.event.details)
            }
          }
        }
      } catch (err: any) {
        if (isActive && err.name !== 'AbortError') {
          console.error('Stream error:', err)
          // エラー時は再接続を試みる
          setTimeout(() => {
            if (isActive) {
              subscribeToEvents()
            }
          }, 3000)
        }
      }
    }

    subscribeToEvents()

    return () => {
      isActive = false
      if (abortController) {
        abortController.abort()
      }
    }
  }, [gameState.gameId, currentPlayerId, onGameStateUpdate])

  const handleEndTurn = async () => {
    try {
      const response = await gameClient.endTurn({
        gameId: gameState.gameId,
        playerId: currentPlayerId,
      })

      if (response.success && response.gameState) {
        onGameStateUpdate(response.gameState)
        setMessage(response.message)
        setSelectedCardId(null)
        setSelectedUnitId(null)
      }
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'ターン終了に失敗しました')
    }
  }

  const handlePlayCard = async (cardId: string, targetId?: string) => {
    try {
      const response = await gameClient.playCard({
        gameId: gameState.gameId,
        playerId: currentPlayerId,
        cardId,
        targetId,
      })

      if (response.success && response.gameState) {
        onGameStateUpdate(response.gameState)
        setMessage(response.message)
        setSelectedCardId(null)
      } else {
        setMessage(response.message || 'カードのプレイに失敗しました')
      }
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'カードのプレイに失敗しました')
    }
  }

  const handleAttack = async (attackerId: string, targetId?: string) => {
    try {
      const response = await gameClient.executeAttack({
        gameId: gameState.gameId,
        playerId: currentPlayerId,
        attackerId,
        targetId,
      })

      if (response.success && response.gameState) {
        onGameStateUpdate(response.gameState)
        setMessage(response.message)
        setSelectedUnitId(null)
      } else {
        setMessage(response.message || '攻撃に失敗しました')
      }
    } catch (err) {
      setMessage(err instanceof Error ? err.message : '攻撃に失敗しました')
    }
  }

  const handleUnitClick = (unitId: string) => {
    if (!isCurrentPlayerTurn) return

    // 自分のユニットをクリックした場合
    const isMyUnit = currentPlayer?.field.some((u) => u.instanceId === unitId)
    if (isMyUnit) {
      setSelectedUnitId(unitId === selectedUnitId ? null : unitId)
      setSelectedCardId(null)
    }
    // 攻撃対象として相手のユニットをクリックした場合
    else if (selectedUnitId) {
      handleAttack(selectedUnitId, unitId)
    }
  }

  const handleOpponentPlayerClick = () => {
    if (!isCurrentPlayerTurn || !selectedUnitId) return
    // プレイヤーへの直接攻撃
    handleAttack(selectedUnitId)
  }

  if (gameState.isGameOver) {
    return (
      <div className="game-over">
        <h2>{gameState.isDraw ? '引き分け' : `${gameState.winnerId === currentPlayerId ? '勝利' : '敗北'}！`}</h2>
        <button onClick={() => window.location.reload()}>新しいゲーム</button>
      </div>
    )
  }

  return (
    <div className="game-board">
      {message && <div className="message-banner">{message}</div>}

      {/* 相手の情報 */}
      <div className="opponent-area">
        <PlayerInfo
          player={opponent!}
          isCurrentPlayer={!isCurrentPlayerTurn}
          onClick={handleOpponentPlayerClick}
          isAttackTarget={!!selectedUnitId}
        />
        <div className="field">
          {opponent?.field.map((unit) => (
            <UnitCard
              key={unit.instanceId}
              unit={unit}
              onClick={() => {
                // カード選択中ならカード使用の対象として渡す
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
          <button className="end-turn-button" onClick={handleEndTurn}>
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
                // カード選択中ならカード使用の対象として渡す（自分のユニットにも対応）
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
        <PlayerInfo 
          player={currentPlayer!} 
          isCurrentPlayer={isCurrentPlayerTurn}
          onClick={() => {}}
          isAttackTarget={false}
        />
        <div className="hand-cards">
            {currentPlayer?.hand && currentPlayer.hand.length > 0 ? (
              currentPlayer.hand.map((card) => (
                <div
                  key={card.id}
                  className={`hand-card-detailed ${selectedCardId === card.id ? 'selected' : ''}`}
                  onClick={() => setSelectedCardId(selectedCardId === card.id ? null : card.id)}
                  title={`${card.name} - コスト:${card.cost} ${card.effect}`}
                >
                  <div className="card-cost">{card.cost}</div>
                  <div className="card-name">{card.name}</div>
                  {/* 魔法カードやユニットカードの効果説明を表示 */}
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
                          {traitLabels[trait] || trait}
                        </span>
                      ))}
                    </div>
                  )}
                  <div className="card-play-button" onClick={(e) => {
                    e.stopPropagation()
                    handlePlayCard(card.id)
                  }}>
                    使用
                  </div>
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


import { useEffect, useRef, useState } from 'react'
import type { GameState } from '../gen/common_pb'
import './CoinToss.css'

interface CoinTossProps {
  gameState: GameState
  playerId: string
  onChooseTurnOrder: (chooseFirst: boolean) => Promise<void>
  coinTossResult?: {
    isHeads: boolean
    winnerId: string
  }
  isFlipping: boolean
  isStopping: boolean
  showResult: boolean
}

export default function CoinToss({
  gameState,
  playerId,
  onChooseTurnOrder,
  coinTossResult,
  isFlipping,
  isStopping,
  showResult,
}: CoinTossProps) {
  const [selectedOrder, setSelectedOrder] = useState<boolean | null>(null)
  const [rotation, setRotation] = useState(0)
  const coinRef = useRef<HTMLDivElement>(null)
  const animationRef = useRef<number>()
  const accumulatedTimeRef = useRef<number>(0)

  // 最初から減速アニメーション
  useEffect(() => {
    console.log('CoinToss useEffect:', {
      isFlipping,
      isStopping,
      coinTossResult,
      hidden: document.hidden,
    })
    if ((isFlipping || isStopping) && coinTossResult) {
      console.log('Starting coin animation')
      accumulatedTimeRef.current = 0

      // 結果に応じて最終角度を決定
      // 表: 3回転(1080度)で0度 → 1080度
      // 裏: 2.5回転(900度)で180度正面 → 900度
      const totalRotation = coinTossResult.isHeads ? 1080 : 900
      const finalAngle = coinTossResult.isHeads ? 0 : 180

      const duration = 3000 // 3秒
      let lastTime: number | null = null

      const animate = (currentTime: number) => {
        if (lastTime === null) {
          lastTime = currentTime
        }

        // タブがアクティブな時だけ時間を加算
        if (!document.hidden) {
          const deltaTime = currentTime - lastTime
          accumulatedTimeRef.current += deltaTime
        }
        lastTime = currentTime

        const progress = Math.min(accumulatedTimeRef.current / duration, 1)

        // ease-outカーブ
        const easeOut = 1 - (1 - progress) ** 3

        const currentRotation = totalRotation * easeOut

        setRotation(currentRotation)

        if (progress < 1) {
          animationRef.current = requestAnimationFrame(animate)
        } else {
          setRotation(finalAngle)
        }
      }

      animationRef.current = requestAnimationFrame(animate)

      return () => {
        if (animationRef.current) {
          cancelAnimationFrame(animationRef.current)
        }
      }
    } else if (showResult && coinTossResult) {
      // 結果表示時は固定する
      setRotation(coinTossResult.isHeads ? 0 : 180)
    }
  }, [isFlipping, isStopping, showResult, coinTossResult])

  const isWinner = gameState.coinTossWinnerId === playerId
  const showTurnOrderChoice =
    gameState.coinTossDone && isWinner && !gameState.turnOrderDecided

  const handleChooseTurnOrder = async (chooseFirst: boolean) => {
    setSelectedOrder(chooseFirst)
    try {
      await onChooseTurnOrder(chooseFirst)
    } catch (error) {
      setSelectedOrder(null)
      console.error('ターン順選択エラー:', error)
    }
  }

  // コイントス中（アニメーション表示）、減速中、または結果表示中
  if (isFlipping || isStopping || showResult || !gameState.coinTossDone) {
    return (
      <div className="coin-toss-container">
        <div className="coin-toss-card">
          <h2>🪙 コイントス</h2>
          <p className="description">
            先攻後攻を決めるためにコイントスを行っています...
          </p>

          <div className="coin-wrapper">
            <div
              ref={coinRef}
              className="coin"
              style={{ transform: `rotateY(${rotation}deg)` }}
            >
              <div className="coin-side heads">表</div>
              <div className="coin-side tails">
                <span className="tails-text-rotating">裏</span>
              </div>
            </div>
          </div>

          <div className="status-message">コイントス中...</div>
        </div>
      </div>
    )
  }

  // ターン順選択（勝者のみ）
  if (showTurnOrderChoice) {
    return (
      <div className="coin-toss-container">
        <div className="coin-toss-card">
          <h2>🎯 先攻後攻を選択</h2>
          <p className="description winner-message">
            おめでとうございます！コイントスに勝利しました
            <br />
            先攻か後攻かを選んでください
          </p>

          {coinTossResult && (
            <div className="coin-result">
              <div
                className={`coin-static ${coinTossResult.isHeads ? 'heads' : 'tails'}`}
              >
                {coinTossResult.isHeads ? (
                  '表'
                ) : (
                  <span className="tails-text">裏</span>
                )}
              </div>
            </div>
          )}

          <div className="turn-order-buttons">
            <button
              type="button"
              className={`turn-order-button first ${selectedOrder === true ? 'selected' : ''}`}
              onClick={() => handleChooseTurnOrder(true)}
              disabled={selectedOrder !== null}
            >
              <span className="button-icon">⚔️</span>
              <span className="button-text">先攻を選ぶ</span>
              <span className="button-desc">最初に攻撃できる</span>
            </button>

            <button
              type="button"
              className={`turn-order-button second ${selectedOrder === false ? 'selected' : ''}`}
              onClick={() => handleChooseTurnOrder(false)}
              disabled={selectedOrder !== null}
            >
              <span className="button-icon">🛡️</span>
              <span className="button-text">後攻を選ぶ</span>
              <span className="button-desc">準備時間がある</span>
            </button>
          </div>
        </div>
      </div>
    )
  }

  // ターン順選択待ち（敗者）
  if (gameState.coinTossDone && !isWinner && !gameState.turnOrderDecided) {
    const winner =
      gameState.player1?.id === gameState.coinTossWinnerId
        ? gameState.player1
        : gameState.player2

    return (
      <div className="coin-toss-container">
        <div className="coin-toss-card">
          <h2>⏳ ターン順決定待ち</h2>
          <p className="description">
            {winner?.name}さんがコイントスに勝利しました
            <br />
            先攻後攻の選択を待っています...
          </p>

          {coinTossResult && (
            <div className="coin-result">
              <div
                className={`coin-static ${coinTossResult.isHeads ? 'heads' : 'tails'}`}
              >
                {coinTossResult.isHeads ? (
                  '表'
                ) : (
                  <span className="tails-text-static">裏</span>
                )}
              </div>
            </div>
          )}

          <div className="waiting-spinner"></div>
        </div>
      </div>
    )
  }

  return null
}

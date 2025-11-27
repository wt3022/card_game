import { useEffect, useRef, useState } from 'react'
import type { GameState } from '../gen/common_pb'
import { useDeckList } from '../hooks/useDeckList'
import { useGameState } from '../hooks/useGameState'
import { useMatchmaking } from '../hooks/useMatchmaking'
import { useMulligan } from '../hooks/useMulligan'
import { gameClient } from '../lib/api-client'
import { getUserInfo } from '../lib/auth'
import type { GameSetupProps } from '../types/components'
import { getMulliganHand, getOpponent } from '../utils/gameHelpers'
import CoinToss from './CoinToss'
import ConnectionStatus from './ConnectionStatus'
import MulliganModal from './MulliganModal'
import './GameSetup.css'

export default function GameSetup({ onGameStart }: GameSetupProps) {
  const [playerName, setPlayerName] = useState('')
  const [selectedDeckId, setSelectedDeckId] = useState<string>('')
  const [showCoinToss, setShowCoinToss] = useState(false)
  const [showMulligan, setShowMulligan] = useState(false)
  const [gameState, setGameState] = useState<GameState | null>(null)
  const [playerId, setPlayerId] = useState('')
  const [_isPerformingCoinToss, setIsPerformingCoinToss] = useState(false)
  const [isCoinFlipping, setIsCoinFlipping] = useState(false)
  const [isCoinStopping, setIsCoinStopping] = useState(false)
  const [showCoinResult, setShowCoinResult] = useState(false)
  const [coinTossResult, setCoinTossResult] = useState<
    | {
        isHeads: boolean
        winnerId: string
      }
    | undefined
  >(undefined)

  // 最新のgameStateを参照するためのref
  const gameStateRef = useRef<GameState | null>(null)

  // ユーザー情報取得
  const userInfo = getUserInfo()

  console.log(
    'GameSetupがレンダリングされました - showMulligan:',
    showMulligan,
    'gameState:',
    gameState?.gameId,
  )

  // マッチング管理
  const {
    isMatchmaking,
    matchmakingStatus,
    error: matchmakingError,
    joinMatchmaking,
    cancelMatchmaking,
  } = useMatchmaking()

  // マリガン管理
  const {
    performMulligan,
    isWaitingForOpponent,
    error: mulliganError,
  } = useMulligan()

  // デッキ一覧取得（ユーザーIDを渡す）
  const { decks, isLoading: isLoadingDecks } = useDeckList(userInfo?.userId)

  // ゲーム状態の購読
  const { gameState: liveGameState } = useGameState(
    gameState?.gameId || null,
    playerId || null,
  )

  // ライブ更新があれば使用
  useEffect(() => {
    if (liveGameState) {
      console.log('GameSetup - ライブゲーム状態更新を受信しました')
      setGameState(liveGameState)
      gameStateRef.current = liveGameState

      // コイントスが完了した場合、結果を設定
      if (
        liveGameState.coinTossDone &&
        !coinTossResult &&
        playerId &&
        liveGameState.coinTossWinnerId
      ) {
        console.log('ライブ更新でコイントス完了を検知しました')
        setCoinTossResult({
          isHeads: liveGameState.coinTossWinnerId === playerId,
          winnerId: liveGameState.coinTossWinnerId,
        })
        setIsCoinFlipping(false)
        setIsCoinStopping(true)

        // 減速アニメーションの時間(3秒)を待つ
        setTimeout(() => {
          setIsCoinStopping(false)
          setShowCoinResult(true)

          // 結果を表示する時間を2秒待つ
          setTimeout(() => {
            setShowCoinResult(false)
          }, 2000)
        }, 3000)
      }

      // コイントス完了＆ターン順決定済み → マリガンへ（3秒後に遷移）
      if (
        liveGameState.coinTossDone &&
        liveGameState.turnOrderDecided &&
        !showMulligan
      ) {
        setTimeout(() => {
          setShowCoinToss(false)
          setShowMulligan(true)
        }, 3000)
      }
    }
  }, [liveGameState, showMulligan, coinTossResult, playerId])

  // エラーメッセージの統合
  const error = matchmakingError || mulliganError

  /**
   * マッチング開始
   */
  const handleJoinMatchmaking = async () => {
    await joinMatchmaking(
      playerName,
      async (matchedGameState, matchedPlayerId) => {
        setGameState(matchedGameState)
        gameStateRef.current = matchedGameState
        setPlayerId(matchedPlayerId)
        setShowCoinToss(true)

        // 画面表示を待つ
        await new Promise((resolve) => setTimeout(resolve, 1000))

        // 最新のgameStateを確認（ライブ更新で既に完了している可能性がある）
        const currentState = gameStateRef.current || matchedGameState

        // コイントスを自動実行（まだ完了していない場合のみ）
        if (!currentState.coinTossDone) {
          setIsPerformingCoinToss(true)
          setIsCoinFlipping(true)

          try {
            const response = await gameClient.performCoinToss({
              gameId: matchedGameState.gameId,
              playerId: matchedPlayerId,
            })

            if (response.success) {
              // 結果を設定
              setCoinTossResult({
                isHeads: response.isHeads,
                winnerId: response.winnerId,
              })
              if (response.gameState) {
                setGameState(response.gameState)
                gameStateRef.current = response.gameState
              }
              setIsCoinFlipping(false)
              setIsCoinStopping(true)

              // 減速アニメーションの時間(3秒)を待つ
              await new Promise((resolve) => setTimeout(resolve, 3000))
              setIsCoinStopping(false)
              setShowCoinResult(true)

              // 結果を表示する時間を2秒待つ
              await new Promise((resolve) => setTimeout(resolve, 2000))
              setShowCoinResult(false)
            }
          } catch (coinTossError: unknown) {
            console.error('コイントスエラー:', coinTossError)
            // 「既に完了」エラーの場合は、ライブ更新で結果が来るのを待つ
            if (
              coinTossError instanceof Error &&
              coinTossError.message?.includes('既に完了')
            ) {
              setIsCoinFlipping(false)
              setIsCoinStopping(false)
            }
          } finally {
            setIsPerformingCoinToss(false)
          }
        } else {
          // 既に完了している場合は、結果を表示
          console.log('コイントスは既に完了しています。結果を表示します。')
          if (currentState.coinTossWinnerId) {
            setCoinTossResult({
              isHeads: currentState.coinTossWinnerId === matchedPlayerId,
              winnerId: currentState.coinTossWinnerId,
            })
            setShowCoinResult(true)
          }
        }
      },
      selectedDeckId || undefined,
    )
  }

  /**
   * ターン順選択
   */
  const handleChooseTurnOrder = async (chooseFirst: boolean) => {
    if (!gameState || !playerId) return

    try {
      const response = await gameClient.chooseTurnOrder({
        gameId: gameState.gameId,
        playerId: playerId,
        chooseFirst: chooseFirst,
      })

      if (response.success && response.gameState) {
        setGameState(response.gameState)
        setShowCoinToss(false)
        setShowMulligan(true)
      }
    } catch (error) {
      console.error('ターン順選択エラー:', error)
    }
  }

  /**
   * マリガン実行
   */
  const handleMulliganSubmit = async (selectedCardIds: string[]) => {
    if (!gameState || !playerId) return

    await performMulligan(
      gameState.gameId,
      playerId,
      selectedCardIds,
      (finalGameState) => {
        onGameStart(finalGameState, playerId)
      },
    )
  }

  /**
   * マリガンスキップ
   */
  const handleMulliganSkip = async () => {
    await handleMulliganSubmit([])
  }

  /**
   * Enterキーでマッチング開始
   */
  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && playerName.trim()) {
      handleJoinMatchmaking()
    }
  }

  // コイントス画面
  if (showCoinToss && gameState) {
    return (
      <CoinToss
        gameState={gameState}
        playerId={playerId}
        onChooseTurnOrder={handleChooseTurnOrder}
        coinTossResult={coinTossResult}
        isFlipping={isCoinFlipping}
        isStopping={isCoinStopping}
        showResult={showCoinResult}
      />
    )
  }

  // マリガン画面
  if (showMulligan && gameState) {
    const opponent = getOpponent(gameState, playerId)

    return (
      <>
        <ConnectionStatus opponent={opponent} />
        <MulliganModal
          hand={getMulliganHand(gameState, playerId)}
          onSubmit={handleMulliganSubmit}
          onSkip={handleMulliganSkip}
          isWaitingForOpponent={isWaitingForOpponent}
        />
      </>
    )
  }

  // マッチング画面
  return (
    <div className="game-setup">
      <div className="setup-card">
        {!isMatchmaking ? (
          <>
            <h2>🎮 対戦相手を探す</h2>
            <p className="description">
              名前を入力してマッチングを開始しましょう
            </p>

            <div className="form-group">
              <label htmlFor="playerName">あなたの名前</label>
              <input
                id="playerName"
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="名前を入力してください"
              />
            </div>

            <div className="form-group">
              <label htmlFor="deckSelect">使用するデッキ</label>
              <select
                id="deckSelect"
                value={selectedDeckId}
                onChange={(e) => setSelectedDeckId(e.target.value)}
                disabled={isLoadingDecks}
              >
                <option value="">デフォルトデッキ (Fixture)</option>
                {decks.map((deck) => (
                  <option key={deck.id} value={deck.id}>
                    {deck.name} ({deck.cardIds.length}枚)
                  </option>
                ))}
              </select>
              {isLoadingDecks && (
                <p className="loading-text">デッキ読み込み中...</p>
              )}
            </div>

            {error && <div className="error-message">{error}</div>}

            <button
              type="button"
              className="start-button"
              onClick={handleJoinMatchmaking}
              disabled={!playerName.trim()}
            >
              マッチング開始
            </button>
          </>
        ) : (
          <>
            <h2>⏳ マッチング中...</h2>
            <div className="matchmaking-status">
              <div className="spinner"></div>
              <p>{matchmakingStatus}</p>
            </div>

            {error && <div className="error-message">{error}</div>}

            <button
              type="button"
              className="cancel-button"
              onClick={cancelMatchmaking}
            >
              キャンセル
            </button>
          </>
        )}
      </div>
    </div>
  )
}

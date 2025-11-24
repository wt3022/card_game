import { useState } from 'react'
import GameBoard from '../components/GameBoard'
import GameSetup from '../components/GameSetup'
import type { GameState } from '../gen/common_pb'
import './Game.css'

function Game() {
  const [gameState, setGameState] = useState<GameState | null>(null)
  const [currentPlayerId, setCurrentPlayerId] = useState<string>('')

  const handleGameStart = (state: GameState, playerId: string) => {
    setGameState(state)
    setCurrentPlayerId(playerId)
  }

  const handleGameStateUpdate = (state: GameState) => {
    setGameState(state)
  }

  return (
    <div className="game-page">
      <div className="app">
        {!gameState ? (
          <GameSetup onGameStart={handleGameStart} />
        ) : (
          <GameBoard
            gameState={gameState}
            currentPlayerId={currentPlayerId}
            onGameStateUpdate={handleGameStateUpdate}
          />
        )}
      </div>
    </div>
  )
}

export default Game

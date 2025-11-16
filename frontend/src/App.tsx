import { useState } from 'react'
import GameSetup from './components/GameSetup'
import GameBoard from './components/GameBoard'
import type { GameState } from './gen/common_pb'
import './App.css'

function App() {
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
  )
}

export default App


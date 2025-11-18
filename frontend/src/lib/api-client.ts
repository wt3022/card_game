import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { GameService } from '../gen/game_connect'

// 環境に応じたベースURLを決定
// 本番環境: nginx経由の api.release-notifier.net を使用
// 開発環境: localhost:8080 に直接アクセス
const getBaseUrl = () => {
  // 開発環境の判定（localhost, 127.0.0.1, またはポート番号がある場合）
  const isDev = window.location.hostname === 'localhost' || 
                window.location.hostname === '127.0.0.1' ||
                window.location.port !== ''
  
  if (isDev) {
    return 'http://localhost:8080'
  }
  
  // 本番環境: nginx でプロキシされた api サブドメインを使用
  return 'https://api.release-notifier.net'
}

// Connect-Webのトランスポートを作成
const transport = createConnectTransport({
  baseUrl: getBaseUrl(),
})

// GameServiceクライアントを作成

export const gameClient = createPromiseClient(GameService, transport)

// StartTurn API呼び出し用ラッパー関数
export async function startTurn({ gameId, playerId }: { gameId: string; playerId: string }) {
  return await gameClient.startTurn({ gameId, playerId })
}


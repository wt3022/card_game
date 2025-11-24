/// <reference types="vite/client" />
import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { GameService } from '../gen/game_connect'

// Viteの推奨パターンでAPIベースURLを取得
const baseUrl = import.meta.env.VITE_API_BASE_URL as string

// Connect-Webのトランスポートを作成
const transport = createConnectTransport({
  baseUrl,
})

// GameServiceクライアントを作成
export const gameClient = createPromiseClient(GameService, transport)

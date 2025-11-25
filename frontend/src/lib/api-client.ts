/// <reference types="vite/client" />

import type { ServiceType } from '@bufbuild/protobuf'
import type { Interceptor, PromiseClient, Transport } from '@connectrpc/connect'
import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { AuthService } from '../gen/auth_connect'
import { CardManagementService } from '../gen/card_management_connect'
import { GameService } from '../gen/game_connect'

// Viteの推奨パターンでAPIベースURLを取得
const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

/**
 * 認証トークンを取得する関数
 */
const getAuthToken = (): string | null => {
  return localStorage.getItem('auth_token')
}

/**
 * 認証インターセプター
 */
const authInterceptor: Interceptor = (next) => async (req) => {
  const token = getAuthToken()
  if (token) {
    req.header.set('Authorization', `Bearer ${token}`)
  }
  
  try {
    const response = await next(req)
    return response
  } catch (error: any) {
    // 認証エラーの場合、トークンをクリアしてログインページにリダイレクト
    if (error?.code === 'unauthenticated' || error?.message?.includes('token is expired')) {
      console.error('Authentication error:', error.message)
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      window.location.href = '/'
    }
    throw error
  }
}

/**
 * 基本トランスポートを作成（認証なし）
 */
const createBaseTransport = (): Transport => {
  return createConnectTransport({
    baseUrl,
  })
}

/**
 * 認証付きトランスポートを作成
 */
const createAuthTransport = (): Transport => {
  return createConnectTransport({
    baseUrl,
    interceptors: [authInterceptor],
  })
}

/**
 * 認証なしでクライアントを作成
 */
export const createClient = <T extends ServiceType>(
  service: T,
): PromiseClient<T> => {
  return createPromiseClient(service, createBaseTransport())
}

/**
 * 認証付きでクライアントを作成
 */
export const createAuthenticatedClient = <T extends ServiceType>(
  service: T,
): PromiseClient<T> => {
  return createPromiseClient(service, createAuthTransport())
}

// 各サービスのクライアントインスタンス
export const gameClient = createClient(GameService)
export const authClient = createClient(AuthService)
export const cardManagementClient = createAuthenticatedClient(
  CardManagementService,
)

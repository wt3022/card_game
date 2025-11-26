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
 * エラーコードの定義
 */
const AUTH_ERROR_CODES = [
  'unauthenticated',
  'permission_denied',
  'invalid_argument',
] as const

const RATE_LIMIT_ERROR_CODES = ['resource_exhausted', 'unavailable'] as const

/**
 * 認証トークンのバリデーション
 */
const isValidToken = (token: string): boolean => {
  if (!token || token.trim().length === 0) {
    return false
  }
  // JWTの基本的な形式チェック（3つのパートがドットで区切られている）
  const parts = token.split('.')
  return parts.length === 3
}

/**
 * 認証インターセプター
 */
const authInterceptor: Interceptor = (next) => async (req) => {
  const token = getAuthToken()
  if (token && isValidToken(token)) {
    req.header.set('Authorization', `Bearer ${token}`)
  } else if (token) {
    // 無効なトークンの場合はクリア
    console.warn('無効な認証トークンを検出しました。クリアします。')
    localStorage.removeItem('auth_token')
    localStorage.removeItem('user_info')
    sessionStorage.removeItem('auth_token')
    sessionStorage.removeItem('user_info')
  }

  try {
    const response = await next(req)
    return response
  } catch (err) {
    const error = err as { code?: string; message?: string }
    // 認証エラーの場合、トークンをクリアしてログインページにリダイレクト
    const isAuthError =
      (error.code &&
        AUTH_ERROR_CODES.includes(
          error.code as (typeof AUTH_ERROR_CODES)[number],
        )) ||
      error.message?.includes('token is expired') ||
      error.message?.includes('invalid token')

    const isRateLimitError =
      error.code &&
      RATE_LIMIT_ERROR_CODES.includes(
        error.code as (typeof RATE_LIMIT_ERROR_CODES)[number],
      )

    if (isAuthError) {
      console.error('認証エラー:', error.message)
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      sessionStorage.removeItem('auth_token')
      sessionStorage.removeItem('user_info')
      // リダイレクトの前に少し待機（ユーザーがエラーを認識できるように）
      setTimeout(() => {
        window.location.href = '/'
      }, 1000)
    } else if (isRateLimitError) {
      console.warn('レート制限超過:', error.message)
      // レート制限エラーはリトライ可能なのでリダイレクトしない
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

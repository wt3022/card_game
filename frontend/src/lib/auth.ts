/// <reference types="vite/client" />
import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { AuthService } from '../gen/auth_connect'

// Viteの推奨パターンでAPIベースURLを取得
const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// 認証付きトランスポートを作成
const createAuthenticatedTransport = (token?: string) => {
  return createConnectTransport({
    baseUrl,
    interceptors: [
      (next) => async (req) => {
        const authToken = token || getToken()
        if (authToken) {
          req.header.set('Authorization', `Bearer ${authToken}`)
        }
        return await next(req)
      },
    ],
  })
}

// トークン管理
const TOKEN_KEY = 'auth_token'
const USER_KEY = 'user_info'

export interface UserInfo {
  userId: string
  username: string
  role: string
}

// トークンを取得
export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY)
}

// トークンを保存
export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token)
}

// トークンを削除
export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

// ユーザー情報を取得
export const getUserInfo = (): UserInfo | null => {
  const userStr = localStorage.getItem(USER_KEY)
  if (!userStr) return null
  try {
    return JSON.parse(userStr)
  } catch {
    return null
  }
}

// ユーザー情報を保存
export const setUserInfo = (userInfo: UserInfo): void => {
  localStorage.setItem(USER_KEY, JSON.stringify(userInfo))
}

// ログイン
export const login = async (
  username: string,
  password: string,
): Promise<UserInfo> => {
  const transport = createConnectTransport({ baseUrl })
  const authClient = createPromiseClient(AuthService, transport)

  const response = await authClient.login({
    username,
    password,
  })

  // トークンとユーザー情報を保存
  setToken(response.accessToken)
  const userInfo: UserInfo = {
    userId: response.userId,
    username: response.username,
    role: response.role,
  }
  setUserInfo(userInfo)

  return userInfo
}

// ログアウト
export const logout = (): void => {
  removeToken()
}

// 認証済みかどうか
export const isAuthenticated = (): boolean => {
  return getToken() !== null
}

// 認証付きAPIクライアントを作成するヘルパー
export function createAuthenticatedClient<
  S extends Parameters<typeof createPromiseClient>[0],
>(Service: S) {
  return createPromiseClient(Service, createAuthenticatedTransport())
}

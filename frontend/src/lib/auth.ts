/// <reference types="vite/client" />
import { authClient } from './api-client'

// ストレージキー定数
const TOKEN_KEY = 'auth_token'
const USER_KEY = 'user_info'

/**
 * ユーザー情報の型定義
 */
export interface UserInfo {
  userId: string
  username: string
  role: string
}

/**
 * トークンを取得
 */
export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY)
}

/**
 * トークンを保存
 */
export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token)
}

/**
 * トークンを削除
 */
export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

/**
 * ユーザー情報を取得
 */
export const getUserInfo = (): UserInfo | null => {
  const userStr = localStorage.getItem(USER_KEY)
  if (!userStr) return null
  try {
    return JSON.parse(userStr)
  } catch {
    return null
  }
}

/**
 * ユーザー情報を保存
 */
export const setUserInfo = (userInfo: UserInfo): void => {
  localStorage.setItem(USER_KEY, JSON.stringify(userInfo))
}

/**
 * ログイン処理
 */
export const login = async (
  username: string,
  password: string,
): Promise<UserInfo> => {
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

/**
 * ログアウト処理
 */
export const logout = (): void => {
  removeToken()
}

/**
 * 認証済みかどうかを確認
 */
export const isAuthenticated = (): boolean => {
  return getToken() !== null
}

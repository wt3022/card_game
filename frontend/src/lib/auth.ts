/// <reference types="vite/client" />
import { authClient } from './api-client'

// ストレージキー定数
const TOKEN_KEY = 'auth_token'
const USER_KEY = 'user_info'

// セキュリティ定数
const MAX_USERNAME_LENGTH = 50
const MIN_PASSWORD_LENGTH = 8
const MAX_PASSWORD_LENGTH = 128

// ストレージの選択：sessionStorageを優先（タブ閉じで自動削除）
// 環境変数でlocalStorageを使うか選択可能
const USE_SESSION_STORAGE = import.meta.env.VITE_USE_SESSION_STORAGE !== 'false'
const storage = USE_SESSION_STORAGE ? sessionStorage : localStorage

/**
 * ユーザー情報の型定義
 */
export interface UserInfo {
  userId: string
  username: string
  role: string
}

/**
 * 認証エラーの型定義
 */
export class AuthError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'AuthError'
  }
}

/**
 * トークンを取得
 */
export const getToken = (): string | null => {
  return storage.getItem(TOKEN_KEY)
}

/**
 * トークンを保存
 */
export const setToken = (token: string): void => {
  storage.setItem(TOKEN_KEY, token)
}

/**
 * トークンを削除
 */
export const removeToken = (): void => {
  storage.removeItem(TOKEN_KEY)
  storage.removeItem(USER_KEY)
  // フォールバック: localStorageからも削除
  if (USE_SESSION_STORAGE) {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }
}

/**
 * ユーザー情報を取得
 */
export const getUserInfo = (): UserInfo | null => {
  const userStr = storage.getItem(USER_KEY)
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
  storage.setItem(USER_KEY, JSON.stringify(userInfo))
}

/**
 * 入力バリデーション
 */
const validateLoginInput = (username: string, password: string): void => {
  if (!username || username.trim().length === 0) {
    throw new AuthError('ユーザー名を入力してください')
  }
  if (username.length > MAX_USERNAME_LENGTH) {
    throw new AuthError(
      `ユーザー名は${MAX_USERNAME_LENGTH}文字以内で入力してください`,
    )
  }
  if (!password || password.length === 0) {
    throw new AuthError('パスワードを入力してください')
  }
  if (password.length < MIN_PASSWORD_LENGTH) {
    throw new AuthError(
      `パスワードは${MIN_PASSWORD_LENGTH}文字以上である必要があります`,
    )
  }
  if (password.length > MAX_PASSWORD_LENGTH) {
    throw new AuthError(
      `パスワードは${MAX_PASSWORD_LENGTH}文字以内で入力してください`,
    )
  }
}

/**
 * レスポンスのバリデーション
 */
const validateLoginResponse = (response: {
  accessToken: string
  userId: string
  username: string
  role: string
}): void => {
  if (!response.accessToken || response.accessToken.trim().length === 0) {
    throw new AuthError('無効な認証トークンを受信しました')
  }
  if (!response.userId || response.userId.trim().length === 0) {
    throw new AuthError('無効なユーザーIDを受信しました')
  }
  if (!response.username || response.username.trim().length === 0) {
    throw new AuthError('無効なユーザー名を受信しました')
  }
}

/**
 * ログイン処理
 */
export const login = async (
  username: string,
  password: string,
): Promise<UserInfo> => {
  // エラー時は必ず既存のトークンをクリア
  try {
    // 入力バリデーション
    validateLoginInput(username, password)

    const response = await authClient.login({
      username: username.trim(),
      password,
    })

    // デバッグ: レスポンスを確認
    console.log('Login response:', {
      hasAccessToken: !!response.accessToken,
      tokenLength: response.accessToken?.length,
      userId: response.userId,
      username: response.username,
    })

    // レスポンスのバリデーション
    validateLoginResponse(response)

    // トークンとユーザー情報を保存
    setToken(response.accessToken)
    const userInfo: UserInfo = {
      userId: response.userId,
      username: response.username,
      role: response.role,
    }
    setUserInfo(userInfo)

    return userInfo
  } catch (error) {
    // エラー発生時は必ずトークンをクリア
    removeToken()

    // エラーハンドリング
    if (error instanceof AuthError) {
      throw error
    }

    // Connect-RPCエラーの処理
    const connectError = error as { code?: string; message?: string }
    if (connectError.code === 'unauthenticated') {
      throw new AuthError('ユーザー名またはパスワードが正しくありません')
    }
    if (connectError.code === 'invalid_argument') {
      throw new AuthError('入力内容に誤りがあります')
    }
    if (connectError.code === 'unavailable') {
      throw new AuthError(
        'サーバーに接続できません。しばらく待ってから再度お試しください',
      )
    }

    if (error instanceof Error) {
      throw new AuthError(`ログインに失敗しました: ${error.message}`)
    }
    throw new AuthError('予期しないエラーが発生しました')
  }
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

import { type ReactNode, useEffect, useState } from 'react'
import { isAuthenticated, logout } from '../../lib/auth'
import LoginForm from './LoginForm'

interface ProtectedRouteProps {
  children: ReactNode
}

export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const [authenticated, setAuthenticated] = useState(isAuthenticated())

  // 認証状態が変更された時にチェック
  useEffect(() => {
    setAuthenticated(isAuthenticated())
  }, [])

  if (!authenticated) {
    return (
      <LoginForm
        onLoginSuccess={() => {
          // トークンが正しく保存されているか再確認
          if (isAuthenticated()) {
            setAuthenticated(true)
          } else {
            // トークンが保存されていない場合はログアウト
            logout()
            setAuthenticated(false)
          }
        }}
      />
    )
  }

  return <>{children}</>
}

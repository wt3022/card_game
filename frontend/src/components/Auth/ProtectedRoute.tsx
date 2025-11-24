import { type ReactNode, useState } from 'react'
import { isAuthenticated } from '../../lib/auth'
import LoginForm from './LoginForm'

interface ProtectedRouteProps {
  children: ReactNode
}

export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const [authenticated, setAuthenticated] = useState(isAuthenticated())

  if (!authenticated) {
    return (
      <LoginForm
        onLoginSuccess={() => {
          setAuthenticated(true)
        }}
      />
    )
  }

  return <>{children}</>
}

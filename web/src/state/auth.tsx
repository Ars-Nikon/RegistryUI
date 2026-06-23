import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { api, type SessionInfo } from '@/lib/api'

interface AuthState {
  session: SessionInfo | null
  loading: boolean
  login: (registryUrl: string, username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<SessionInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const qc = useQueryClient()

  // Restore an existing auth on first load.
  useEffect(() => {
    let active = true
    api
      .session()
      .then((s) => active && setSession(s))
      .catch(() => active && setSession(null))
      .finally(() => active && setLoading(false))
    return () => {
      active = false
    }
  }, [])

  const login = useCallback(
    async (registryUrl: string, username: string, password: string) => {
      const s = await api.login(registryUrl, username, password)
      setSession(s)
      qc.clear()
    },
    [qc],
  )

  const logout = useCallback(async () => {
    await api.logout().catch(() => {})
    setSession(null)
    qc.clear()
  }, [qc])

  const value = useMemo<AuthState>(
    () => ({ session, loading, login, logout }),
    [session, loading, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

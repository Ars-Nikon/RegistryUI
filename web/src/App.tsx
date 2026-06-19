import { Routes, Route, Navigate } from 'react-router-dom'
import { useApp } from '@/state/app'
import { useAuth } from '@/state/auth'
import { Sidebar } from '@/components/Sidebar'
import { Login } from '@/components/Login'
import { CommandPalette } from '@/components/CommandPalette'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { Toasts } from '@/components/Toasts'
import { Catalog } from '@/pages/Catalog'
import { Tags } from '@/pages/Tags'
import { Image } from '@/pages/Image'
import { Settings } from '@/pages/Settings'

function Shell() {
  return (
    <div className="app">
      <Sidebar />
      <main className="main">
        <div className="main-scroll">
          <div className="container">
            <Routes>
              <Route path="/" element={<Catalog />} />
              <Route path="/repository" element={<Tags />} />
              <Route path="/tag" element={<Image />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </div>
        </div>
      </main>
    </div>
  )
}

export default function App() {
  const { theme } = useApp()
  const { session, loading } = useAuth()

  return (
    <div className={`approot${theme === 'dark' ? ' dark' : ''}`}>
      {!loading && (session ? <Shell /> : <Login />)}
      {session && (
        <>
          <CommandPalette />
          <ConfirmDialog />
        </>
      )}
      <Toasts />
    </div>
  )
}

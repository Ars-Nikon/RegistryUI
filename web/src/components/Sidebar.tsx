import { useLocation, useNavigate } from 'react-router-dom'
import { useApp } from '@/state/app'
import { useAuth } from '@/state/auth'
import { useStats } from '@/hooks/queries'
import { fmtBytes } from '@/lib/format'
import { CubeIcon, CubeOutlineIcon, SettingsIcon, SignOutIcon } from '@/components/icons'

export function Sidebar() {
  const { t } = useApp()
  const { session, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const { data: stats } = useStats()

  const onCatalog = location.pathname === '/' || location.pathname.startsWith('/repository') || location.pathname.startsWith('/tag')
  const onSettings = location.pathname.startsWith('/settings')

  let host = session?.registryUrl ?? ''
  try {
    if (session?.registryUrl) host = new URL(session.registryUrl).host
  } catch {
    /* keep raw */
  }
  const user = session?.username || 'anonymous'

  return (
    <aside className="sidebar">
      <div className="sb-head">
        <div className="sb-brand">
          <div className="brand-badge">
            <CubeIcon size={17} />
          </div>
          <div style={{ minWidth: 0 }}>
            <div className="brand-title">RegistryUI</div>
            <div className="brand-host">{host}</div>
          </div>
        </div>
      </div>

      <nav className="nav">
        <button className={`nav-btn${onCatalog ? ' active' : ''}`} onClick={() => navigate('/')}>
          <CubeOutlineIcon size={16} />
          {t.nav_repos}
        </button>
        <button className={`nav-btn${onSettings ? ' active' : ''}`} onClick={() => navigate('/settings')}>
          <SettingsIcon size={16} />
          {t.nav_settings}
        </button>
      </nav>

      <div className="stats-card">
        <div className="stat-row">
          <span className="stat-label">{t.stat_repos}</span>
          <span className="stat-val">{stats?.repositories ?? '—'}</span>
        </div>
        <div className="stat-row">
          <span className="stat-label">{t.stat_tags}</span>
          <span className="stat-val">{stats?.tags ?? '—'}</span>
        </div>
        <div className="stat-row">
          <span className="stat-label">{t.stat_storage}</span>
          <span className="stat-val">{stats ? fmtBytes(stats.storage) : '—'}</span>
        </div>
      </div>

      <div className="sb-spacer" />

      <div className="account">
        <div className="avatar">{user.charAt(0)}</div>
        <div className="account-info">
          <div className="account-name">{user}</div>
          <div className="account-role">push · pull · delete</div>
        </div>
        <button className="icon-btn" title={t.signout} onClick={() => logout()}>
          <SignOutIcon size={15} />
        </button>
      </div>
    </aside>
  )
}

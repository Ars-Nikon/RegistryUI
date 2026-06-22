import { useEffect, useState } from 'react'
import { api, type RegistryOption } from '@/lib/api'
import { useAuth } from '@/state/auth'
import { useApp } from '@/state/app'
import { CubeIcon, LockIcon } from '@/components/icons'

export function Login() {
  const { login } = useAuth()
  const { t, lang, setLang } = useApp()
  const [registries, setRegistries] = useState<RegistryOption[]>([])
  const [registryUrl, setRegistryUrl] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  // Load the allow-listed registries; select the first one. Credentials are
  // never prefilled — the user always enters them.
  useEffect(() => {
    api
      .defaults()
      .then((d) => {
        setRegistries(d.registries)
        if (d.registries.length > 0) setRegistryUrl(d.registries[0].url)
      })
      .catch(() => {})
  }, [])

  const host = (() => {
    try {
      return registryUrl ? new URL(registryUrl).host : 'registry'
    } catch {
      return registryUrl || 'registry'
    }
  })()

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await login(registryUrl, username, password)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="login-wrap">
      <div className="login-outer">
        <div className="login-brand">
          <div className="brand-badge">
            <CubeIcon size={19} />
          </div>
          <span className="login-brand-name">RegistryUI</span>
        </div>
        <form className="login-card" onSubmit={onSubmit}>
          <div className="login-head">
            <div className="login-head-title">{t.login_signin}</div>
            <div className="seg sm">
              <button
                type="button"
                className={`seg-btn${lang === 'en' ? ' active' : ''}`}
                onClick={() => setLang('en')}
              >
                EN
              </button>
              <button
                type="button"
                className={`seg-btn${lang === 'ru' ? ' active' : ''}`}
                onClick={() => setLang('ru')}
              >
                RU
              </button>
            </div>
          </div>
          <div className="login-sub">
            {t.login_connect} <span className="mono">{host}</span>
          </div>

          <label className="field-label">{t.login_url}</label>
          <select
            className="field-input field-select"
            value={registryUrl}
            onChange={(e) => setRegistryUrl(e.target.value)}
          >
            {registries.map((r) => (
              <option key={r.url} value={r.url}>
                {r.name}
              </option>
            ))}
          </select>
          <label className="field-label">{t.login_user}</label>
          <input
            className="field-input"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
          />
          <label className="field-label">{t.login_pass}</label>
          <input
            className="field-input last"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
          />
          <button className="btn-primary" type="submit" disabled={busy || !registryUrl}>
            {busy ? t.loading : t.login_signin}
          </button>

          {error && <div className="login-error">{error}</div>}

          <div className="login-note">
            <LockIcon size={15} style={{ flex: 'none', marginTop: 1 }} />
            <span>{t.login_note}</span>
          </div>
        </form>
      </div>
    </div>
  )
}

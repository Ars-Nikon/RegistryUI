import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTagDetails } from '@/hooks/queries'
import { useApp } from '@/state/app'
import { useAuth } from '@/state/auth'
import type { TagDetails } from '@/lib/api'
import { fmtBytes, fmtDate, shortDigest, joinArgs } from '@/lib/format'
import { CopyIcon, CheckIcon, TrashSimpleIcon } from '@/components/icons'

function registryHost(url: string | undefined): string {
  if (!url) return 'registry'
  try {
    return new URL(url).host
  } catch {
    return url.replace(/^https?:\/\//, '')
  }
}

export function Image() {
  const [params] = useSearchParams()
  const repo = params.get('repo') ?? ''
  const tag = params.get('tag') ?? ''
  const { t, lang, copy, askDelete } = useApp()
  const { session } = useAuth()
  const [selected, setSelected] = useState('')
  const [digestCopied, setDigestCopied] = useState(false)
  const [pullCopied, setPullCopied] = useState(false)

  const { data: main, isLoading, error } = useTagDetails(repo, tag)
  const firstDigest = main?.platforms?.[0]?.digest ?? ''
  const activeDigest = selected || firstDigest
  const { data: child } = useTagDetails(repo, main?.isIndex ? activeDigest : '')

  const view: TagDetails | undefined = main?.isIndex ? child : main
  const host = registryHost(session?.registryUrl)
  const pullRef = `${host}/${repo}:${tag}`

  const maxLayer = useMemo(
    () => Math.max(1, ...(view?.layers ?? []).map((l) => l.size)),
    [view],
  )

  const copyDigest = () => {
    if (main) copy(main.digest)
    setDigestCopied(true)
    setTimeout(() => setDigestCopied(false), 1400)
  }
  const copyPull = () => {
    copy(`docker pull ${pullRef}`)
    setPullCopied(true)
    setTimeout(() => setPullCopied(false), 1400)
  }

  if (isLoading) return <div className="skeleton" style={{ height: 320 }} />
  if (error)
    return (
      <div className="empty">
        <div className="empty-title">{t.load_error}</div>
        <div className="empty-sub">{(error as Error).message}</div>
      </div>
    )
  if (!main) return null

  const labels = view?.labels ? Object.entries(view.labels) : []

  return (
    <div className="fade">
      <div className="img-head">
        <div style={{ minWidth: 0 }}>
          <h1 className="img-title">
            {repo}:{tag}
          </h1>
          <button className="img-digest-btn" onClick={copyDigest}>
            {shortDigest(main.digest)}
            {digestCopied ? (
              <CheckIcon size={13} stroke="var(--green)" strokeWidth={2.4} />
            ) : (
              <CopyIcon size={13} style={{ opacity: 0.6 }} />
            )}
          </button>
        </div>
        <button className="btn-danger-outline" onClick={() => askDelete({ repo, tag, digest: main.digest })}>
          <TrashSimpleIcon size={15} />
          {t.delete_tag}
        </button>
      </div>

      <div className="pull-bar invert">
        <span className="pull-dollar">$</span>
        <code className="pull-code">docker pull {pullRef}</code>
        <button className="pull-copy-accent" onClick={copyPull}>
          {pullCopied ? <CheckIcon size={14} stroke="#fff" strokeWidth={2.6} /> : <CopyIcon size={14} />}
          {pullCopied ? t.copied : t.copy}
        </button>
      </div>

      {main.isIndex && main.platforms && main.platforms.length > 0 && (
        <div className="platforms">
          <div className="platforms-label">{t.platforms_list}</div>
          <div className="platform-list">
            {main.platforms.map((p) => {
              const label = `${p.os}/${p.architecture}${p.variant ? `/${p.variant}` : ''}`
              const active = (selected || firstDigest) === p.digest
              return (
                <button
                  key={p.digest}
                  className={`platform-btn${active ? ' active' : ''}`}
                  onClick={() => setSelected(p.digest)}
                >
                  <span className="platform-name">{label}</span>
                  <span className="platform-size">{fmtBytes(p.size)}</span>
                </button>
              )
            })}
          </div>
        </div>
      )}

      <div className="summary-grid">
        <div className="summary-card">
          <div className="summary-label">{t.total_size}</div>
          <div className="summary-val">{fmtBytes(view?.size ?? main.size)}</div>
        </div>
        <div className="summary-card">
          <div className="summary-label">{t.created}</div>
          <div className="summary-val sm">{fmtDate(view?.created, lang)}</div>
        </div>
        <div className="summary-card">
          <div className="summary-label">{t.platform}</div>
          <div className="summary-val sm mono">
            {view?.os ?? '—'}/{view?.architecture ?? '—'}
          </div>
        </div>
        <div className="summary-card">
          <div className="summary-label">{t.layers}</div>
          <div className="summary-val">{view?.layers?.length ?? 0}</div>
        </div>
      </div>

      <div className="config-card">
        <div className="config-head">{t.configuration}</div>
        <div className="config-body">
          <div className="config-row">
            <span className="config-key">Entrypoint</span>
            <code className="config-val">{joinArgs(view?.entrypoint)}</code>
          </div>
          <div className="config-row">
            <span className="config-key">Cmd</span>
            <code className="config-val">{joinArgs(view?.cmd)}</code>
          </div>
          <div className="config-row">
            <span className="config-key">WorkingDir</span>
            <code className="config-val">{view?.workingDir || '/'}</code>
          </div>
          <div className="config-row tall">
            <span className="config-key">Env</span>
            <div className="kv-list">
              {(view?.env ?? []).length === 0 && <span className="config-val">—</span>}
              {(view?.env ?? []).map((e, i) => {
                const idx = e.indexOf('=')
                const k = idx >= 0 ? e.slice(0, idx) : e
                const v = idx >= 0 ? e.slice(idx + 1) : ''
                return (
                  <div className="kv-item" key={i}>
                    <span className="k-accent">{k}</span>
                    <span className="eq">=</span>
                    <span>{v}</span>
                  </div>
                )
              })}
            </div>
          </div>
          {labels.length > 0 && (
            <div className="config-row tall">
              <span className="config-key">Labels</span>
              <div className="kv-list">
                {labels.map(([k, v]) => (
                  <div className="kv-item" key={k}>
                    <span className="k-muted">{k}</span>
                    <span className="eq">=</span>
                    <span>{v}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="layers-card">
        <div className="layers-head">
          <span>{t.layers}</span>
          <span className="layers-summary">
            {view?.layers?.length ?? 0} · {fmtBytes(view?.size ?? 0)}
          </span>
        </div>
        {(view?.layers ?? []).map((l, i) => (
          <div className="layer-row" key={l.digest + i}>
            <span className="layer-idx">{i + 1}</span>
            <span className="layer-digest">{shortDigest(l.digest)}</span>
            <div>
              <div className="layer-size">{fmtBytes(l.size)}</div>
              <div className="bar">
                <div className="bar-fill" style={{ width: `${(l.size / maxLayer) * 100}%` }} />
              </div>
            </div>
            <code className="layer-by" title={l.createdBy}>
              {l.createdBy || '—'}
            </code>
          </div>
        ))}
      </div>
    </div>
  )
}

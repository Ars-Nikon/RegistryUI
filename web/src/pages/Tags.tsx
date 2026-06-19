import { useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQueries } from '@tanstack/react-query'
import { api, type TagDetails } from '@/lib/api'
import { keys, useRepoSummary, useTags } from '@/hooks/queries'
import { useApp } from '@/state/app'
import { useAuth } from '@/state/auth'
import { fmtBytes, relTime, fmtDate, shortDigest } from '@/lib/format'
import { CopyIcon, CheckIcon, EyeIcon, TrashSimpleIcon, ChevronDownIcon } from '@/components/icons'

type SortKey = 'tag' | 'created' | 'size'
type SortDir = 'asc' | 'desc'

interface Row {
  tag: string
  details?: TagDetails
}

function registryHost(url: string | undefined): string {
  if (!url) return 'registry'
  try {
    return new URL(url).host
  } catch {
    return url.replace(/^https?:\/\//, '')
  }
}

function platformLabel(d: TagDetails | undefined): string {
  if (!d) return '…'
  if (d.isIndex) return `${d.platforms?.length ?? 0} platforms`
  return `${d.os ?? 'linux'}/${d.architecture ?? 'amd64'}`
}

export function Tags() {
  const [params] = useSearchParams()
  const repo = params.get('repo') ?? ''
  const { t, lang, copy, askDelete } = useApp()
  const { session } = useAuth()
  const navigate = useNavigate()

  const { data: tags, isLoading, error } = useTags(repo)
  const { data: summary } = useRepoSummary(repo)
  const [sort, setSort] = useState<{ key: SortKey; dir: SortDir }>({ key: 'created', dir: 'desc' })
  const [pullCopied, setPullCopied] = useState(false)

  const detailQueries = useQueries({
    queries: (tags ?? []).map((tag) => ({
      queryKey: keys.tag(repo, tag),
      queryFn: () => api.tagDetails(repo, tag),
      enabled: !!repo,
    })),
  })

  const host = registryHost(session?.registryUrl)
  const pullRef = `${host}/${repo}`

  const rows = useMemo<Row[]>(() => {
    const base: Row[] = (tags ?? []).map((tag, i) => ({ tag, details: detailQueries[i]?.data }))
    const dir = sort.dir === 'asc' ? 1 : -1
    base.sort((a, b) => {
      if (sort.key === 'tag') return a.tag.localeCompare(b.tag) * dir
      if (sort.key === 'size') return ((a.details?.size ?? 0) - (b.details?.size ?? 0)) * dir
      const ta = a.details?.created ? Date.parse(a.details.created) : 0
      const tb = b.details?.created ? Date.parse(b.details.created) : 0
      return (ta - tb) * dir
    })
    return base
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tags, sort, detailQueries.map((q) => q.dataUpdatedAt).join(',')])

  const toggleSort = (key: SortKey) =>
    setSort((s) => (s.key === key ? { key, dir: s.dir === 'asc' ? 'desc' : 'asc' } : { key, dir: 'desc' }))

  const copyPull = () => {
    copy(`docker pull ${pullRef}`)
    setPullCopied(true)
    setTimeout(() => setPullCopied(false), 1400)
  }

  const chev = (key: SortKey) => (
    <ChevronDownIcon
      size={11}
      strokeWidth={3}
      className="th-chev"
      style={{
        opacity: sort.key === key ? 1 : 0.25,
        transform: sort.key === key && sort.dir === 'asc' ? 'rotate(180deg)' : 'none',
      }}
    />
  )

  const countText = lang === 'ru' ? `${tags?.length ?? 0} тегов` : `${tags?.length ?? 0} tags`

  return (
    <div className="fade">
      <h1 className="tags-title">{repo}</h1>
      <p className="page-sub" style={{ margin: '5px 0 18px' }}>
        {summary?.description || ' '}
      </p>

      <div className="pull-bar">
        <span className="pull-dollar">$</span>
        <code className="pull-code">docker pull {pullRef}</code>
        <button className="copy-btn" onClick={copyPull}>
          {pullCopied ? <CheckIcon size={14} stroke="var(--green)" strokeWidth={2.4} /> : <CopyIcon size={14} />}
          {t.copy}
        </button>
      </div>

      <div className="tags-count">{countText}</div>

      {isLoading && (
        <div className="repo-list">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="skeleton" style={{ height: 56 }} />
          ))}
        </div>
      )}
      {error && (
        <div className="empty">
          <div className="empty-title">{t.load_error}</div>
          <div className="empty-sub">{(error as Error).message}</div>
        </div>
      )}

      {tags && tags.length > 0 && (
        <div className="table">
          <div className="thead">
            <span className="th-sort-group">
              <button className={`th-btn${sort.key === 'tag' ? ' active' : ''}`} onClick={() => toggleSort('tag')}>
                {t.th_tag}
                {chev('tag')}
              </button>
              <button
                className={`th-btn${sort.key === 'created' ? ' active' : ''}`}
                onClick={() => toggleSort('created')}
              >
                {t.th_created}
                {chev('created')}
              </button>
            </span>
            <span className="th-static">{t.th_digest}</span>
            <button className={`th-btn${sort.key === 'size' ? ' active' : ''}`} onClick={() => toggleSort('size')}>
              {t.th_size}
              {chev('size')}
            </button>
            <span className="th-static right">{t.th_actions}</span>
          </div>

          {rows.map(({ tag, details }) => (
            <div
              className="trow"
              key={tag}
              onClick={() => navigate(`/tag?repo=${encodeURIComponent(repo)}&tag=${encodeURIComponent(tag)}`)}
            >
              <div style={{ minWidth: 0 }}>
                <div className="tag-name-row">
                  <span className="tag-name">{tag}</span>
                  {details?.isIndex && <span className="multi-badge">{t.multi_arch}</span>}
                </div>
                <div className="tag-meta">
                  <span className="mono">{platformLabel(details)}</span>
                  <span className="dot">·</span>
                  <span title={details?.created ? fmtDate(details.created, lang) : ''}>
                    {details?.created ? relTime(details.created, lang) : '—'}
                  </span>
                </div>
              </div>
              <button
                className="digest-btn"
                onClick={(e) => {
                  e.stopPropagation()
                  if (details) copy(details.digest)
                }}
                title={t.copy_digest}
              >
                <span>{shortDigest(details?.digest)}</span>
                <CopyIcon size={12} style={{ flex: 'none', opacity: 0.6 }} />
              </button>
              <span className="size-cell">{details ? fmtBytes(details.size) : '…'}</span>
              <div className="row-actions">
                <button
                  className="act-btn"
                  title={t.view_image}
                  onClick={(e) => {
                    e.stopPropagation()
                    navigate(`/tag?repo=${encodeURIComponent(repo)}&tag=${encodeURIComponent(tag)}`)
                  }}
                >
                  <EyeIcon size={14} />
                </button>
                <button
                  className="act-btn danger"
                  title={t.delete_tag}
                  onClick={(e) => {
                    e.stopPropagation()
                    askDelete({ repo, tag, digest: details?.digest ?? '' })
                  }}
                >
                  <TrashSimpleIcon size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

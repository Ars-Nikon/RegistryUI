import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRepositories, useRepoSummary } from '@/hooks/queries'
import { useApp } from '@/state/app'
import { fmtBytes, relTime } from '@/lib/format'
import { CubeIcon, SearchIcon, ChevronRightIcon, PlusIcon } from '@/components/icons'

const PAGE = 20

function RepoCard({ repo }: { repo: string }) {
  const { t, lang } = useApp()
  const navigate = useNavigate()
  const { data } = useRepoSummary(repo)

  return (
    <div
      className="repo-card"
      onClick={() => navigate(`/repository?repo=${encodeURIComponent(repo)}`)}
    >
      <div className="repo-icon">
        <CubeIcon size={20} />
      </div>
      <div className="repo-main">
        <div className="repo-name">{repo}</div>
        <div className="repo-desc">{data?.description || ' '}</div>
      </div>
      <div className="repo-stats">
        <div className="repo-stat" style={{ minWidth: 54 }}>
          <div className="val">{data ? data.tagCount : '—'}</div>
          <div className="lbl">{t.lbl_tags}</div>
        </div>
        <div className="repo-stat" style={{ minWidth: 64 }}>
          <div className="val mono">{data ? fmtBytes(data.size) : '—'}</div>
          <div className="lbl">{t.lbl_size}</div>
        </div>
        <div className="repo-stat" style={{ minWidth: 80 }}>
          <div className="val" style={{ fontWeight: 500 }}>
            {data?.updated ? relTime(data.updated, lang) : '—'}
          </div>
          <div className="lbl">{t.lbl_updated}</div>
        </div>
        <ChevronRightIcon size={18} stroke="var(--text3)" />
      </div>
    </div>
  )
}

export function Catalog() {
  const { t, lang } = useApp()
  const { data: repos, isLoading, error } = useRepositories()
  const [filter, setFilter] = useState('')
  const [shown, setShown] = useState(PAGE)

  const filtered = useMemo(() => {
    const list = repos ?? []
    const q = filter.trim().toLowerCase()
    return q ? list.filter((r) => r.toLowerCase().includes(q)) : list
  }, [repos, filter])

  const visible = filtered.slice(0, shown)
  const hasMore = filtered.length > shown
  const countText =
    lang === 'ru'
      ? `${filtered.length} репозиториев`
      : `${filtered.length} ${filtered.length === 1 ? 'repository' : 'repositories'}`

  return (
    <div className="fade">
      <h1 className="page-title">{t.catalog_title}</h1>
      <p className="page-sub" style={{ margin: '0 0 22px' }}>
        {countText}
      </p>

      <div className="search">
        <SearchIcon size={16} stroke="var(--text3)" className="search-icon" />
        <input
          className="search-input"
          value={filter}
          onChange={(e) => {
            setFilter(e.target.value)
            setShown(PAGE)
          }}
          placeholder={t.filter_ph}
        />
      </div>

      {isLoading && (
        <div className="repo-list">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="skeleton" style={{ height: 72 }} />
          ))}
        </div>
      )}

      {error && (
        <div className="empty">
          <div className="empty-title">{t.load_error}</div>
          <div className="empty-sub">{(error as Error).message}</div>
        </div>
      )}

      {repos && filtered.length === 0 && (
        <div className="empty">
          <div className="empty-icon">
            <SearchIcon size={22} />
          </div>
          <div className="empty-title">{t.empty_title}</div>
          <div className="empty-sub">{t.empty_sub}</div>
          <button className="btn-outline" onClick={() => setFilter('')}>
            {t.clear_filter}
          </button>
        </div>
      )}

      <div className="repo-list">
        {visible.map((repo) => (
          <RepoCard key={repo} repo={repo} />
        ))}
      </div>

      {hasMore && (
        <button className="load-more" onClick={() => setShown((n) => n + PAGE)}>
          <PlusIcon size={15} />
          {filtered.length - shown}
        </button>
      )}
    </div>
  )
}

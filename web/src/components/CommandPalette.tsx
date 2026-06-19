import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useApp } from '@/state/app'
import { useRepositories } from '@/hooks/queries'
import { SearchIcon } from '@/components/icons'

interface Result {
  type: 'repo' | 'tag'
  label: string
  sub: string
  to: string
}

export function CommandPalette() {
  const { paletteOpen, closePalette, t } = useApp()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { data: repos } = useRepositories()
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (paletteOpen) {
      setQuery('')
      requestAnimationFrame(() => inputRef.current?.focus())
    }
  }, [paletteOpen])

  const results = useMemo<Result[]>(() => {
    if (!paletteOpen) return []
    const q = query.trim().toLowerCase()
    const out: Result[] = []
    for (const repo of repos ?? []) {
      if (!q || repo.toLowerCase().includes(q)) {
        out.push({ type: 'repo', label: repo, sub: t.type_repo, to: `/repository?repo=${encodeURIComponent(repo)}` })
      }
      // Tags already loaded into the query cache become searchable too.
      const tags = qc.getQueryData<string[]>(['tags', repo])
      if (tags) {
        for (const tag of tags) {
          if (!q || `${repo}:${tag}`.toLowerCase().includes(q)) {
            out.push({
              type: 'tag',
              label: `${repo}:${tag}`,
              sub: t.type_tag,
              to: `/tag?repo=${encodeURIComponent(repo)}&tag=${encodeURIComponent(tag)}`,
            })
          }
        }
      }
    }
    return out.slice(0, 50)
  }, [paletteOpen, query, repos, qc, t])

  if (!paletteOpen) return null

  const go = (to: string) => {
    closePalette()
    navigate(to)
  }

  return (
    <div className="overlay" onClick={closePalette}>
      <div className="palette" onClick={(e) => e.stopPropagation()}>
        <div className="palette-head">
          <SearchIcon size={18} stroke="var(--text3)" />
          <input
            ref={inputRef}
            className="palette-input"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t.palette_ph}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && results[0]) go(results[0].to)
            }}
          />
          <span className="kbd">esc</span>
        </div>
        <div className="palette-list">
          {results.length === 0 && <div className="palette-empty">{t.no_matches}</div>}
          {results.map((r, i) => (
            <button className="palette-item" key={i} onClick={() => go(r.to)}>
              <span className={`ptype ${r.type}`}>{r.sub}</span>
              <span className="plabel">{r.label}</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

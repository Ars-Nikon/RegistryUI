import type { Lang } from '@/i18n'

/** Human-readable byte size, matching the mockup's rounding rules. */
export function fmtBytes(b: number): string {
  if (!b || b < 0) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = Math.floor(Math.log(b) / Math.log(1024))
  i = Math.max(0, Math.min(i, u.length - 1))
  const v = b / Math.pow(1024, i)
  const s = i === 0 ? Math.round(v) : v >= 100 ? Math.round(v) : v.toFixed(v >= 10 ? 1 : 2)
  return `${s} ${u[i]}`
}

/** Numeric date: "20.06.2026" (ru) / "06/20/2026" (en). */
export function fmtDate(iso: string | undefined, lang: Lang): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  const day = String(d.getDate()).padStart(2, '0')
  const mon = String(d.getMonth() + 1).padStart(2, '0')
  const year = d.getFullYear()
  return lang === 'ru' ? `${day}.${mon}.${year}` : `${mon}/${day}/${year}`
}

/** Truncated digest: "sha256:abc123…" -> "sha256:abc123def456". */
export function shortDigest(d: string | undefined): string {
  if (!d) return '—'
  const hex = d.replace('sha256:', '')
  return 'sha256:' + hex.slice(0, 12)
}

/** Join an argv array into a readable command string. */
export function joinArgs(a: string[] | undefined): string {
  if (!a || a.length === 0) return '—'
  return a.join(' ')
}

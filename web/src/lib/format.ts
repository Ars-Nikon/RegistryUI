import type { Lang } from '@/i18n'

const MONTHS: Record<Lang, string[]> = {
  en: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
  ru: ['янв.', 'февр.', 'марта', 'апр.', 'мая', 'июня', 'июля', 'авг.', 'сент.', 'окт.', 'нояб.', 'дек.'],
}

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

/** Absolute date, localized per language. */
export function fmtDate(iso: string | undefined, lang: Lang): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  const mo = MONTHS[lang]
  return lang === 'ru'
    ? `${d.getDate()} ${mo[d.getMonth()]} ${d.getFullYear()}`
    : `${mo[d.getMonth()]} ${d.getDate()}, ${d.getFullYear()}`
}

/** Relative time ("3d ago" / "3 дн. назад"). */
export function relTime(iso: string | undefined, lang: Lang): string {
  if (!iso) return '—'
  const d = new Date(iso).getTime()
  if (Number.isNaN(d)) return '—'
  const now = Date.now()
  const sec = (now - d) / 1000
  const day = 86400
  const ru = lang === 'ru'
  if (sec < day) return ru ? 'сегодня' : 'today'
  if (sec < 2 * day) return ru ? 'вчера' : 'yesterday'
  const days = Math.floor(sec / day)
  if (days < 30) return ru ? `${days} дн. назад` : `${days}d ago`
  const mo = Math.floor(days / 30)
  if (mo < 12) return ru ? `${mo} мес. назад` : `${mo}mo ago`
  const y = Math.floor(days / 365)
  return ru ? `${y} г. назад` : `${y}y ago`
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

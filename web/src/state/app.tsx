import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { T, type Lang, type Strings } from '@/i18n'

type Theme = 'light' | 'dark'

export interface Toast {
  id: number
  text: string
  color: string
}

export interface ConfirmTarget {
  repo: string
  tag: string
  digest: string
}

interface AppState {
  theme: Theme
  lang: Lang
  t: Strings
  toasts: Toast[]
  paletteOpen: boolean
  confirm: ConfirmTarget | null
  setTheme: (t: Theme) => void
  setLang: (l: Lang) => void
  addToast: (text: string, color?: string) => void
  copy: (text: string) => void
  openPalette: () => void
  closePalette: () => void
  askDelete: (target: ConfirmTarget) => void
  cancelDelete: () => void
}

const AppContext = createContext<AppState | null>(null)

const THEME_KEY = 'rui.theme'
const LANG_KEY = 'rui.lang'

export function AppProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem(THEME_KEY) as Theme) || 'light',
  )
  const [lang, setLangState] = useState<Lang>(
    () => (localStorage.getItem(LANG_KEY) as Lang) || 'en',
  )
  const [toasts, setToasts] = useState<Toast[]>([])
  const [paletteOpen, setPaletteOpen] = useState(false)
  const [confirm, setConfirm] = useState<ConfirmTarget | null>(null)
  const nextId = useRef(1)

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t)
    localStorage.setItem(THEME_KEY, t)
  }, [])
  const setLang = useCallback((l: Lang) => {
    setLangState(l)
    localStorage.setItem(LANG_KEY, l)
  }, [])

  const addToast = useCallback((text: string, color = 'var(--accent)') => {
    const id = nextId.current++
    setToasts((prev) => [...prev, { id, text, color }])
    setTimeout(() => setToasts((prev) => prev.filter((x) => x.id !== id)), 3600)
  }, [])

  const copy = useCallback(
    (text: string) => {
      navigator.clipboard?.writeText(text)
      addToast(T[lang].toast_copied, 'var(--green)')
    },
    [addToast, lang],
  )

  const openPalette = useCallback(() => setPaletteOpen(true), [])
  const closePalette = useCallback(() => setPaletteOpen(false), [])
  const askDelete = useCallback((target: ConfirmTarget) => setConfirm(target), [])
  const cancelDelete = useCallback(() => setConfirm(null), [])

  // ⌘K / Ctrl+K opens the command palette anywhere.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        setPaletteOpen((v) => !v)
      }
      if (e.key === 'Escape') {
        setPaletteOpen(false)
        setConfirm(null)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [])

  const value = useMemo<AppState>(
    () => ({
      theme,
      lang,
      t: T[lang],
      toasts,
      paletteOpen,
      confirm,
      setTheme,
      setLang,
      addToast,
      copy,
      openPalette,
      closePalette,
      askDelete,
      cancelDelete,
    }),
    [theme, lang, toasts, paletteOpen, confirm, setTheme, setLang, addToast, copy, openPalette, closePalette, askDelete, cancelDelete],
  )

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useApp() {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useApp must be used within AppProvider')
  return ctx
}

import { useApp } from '@/state/app'

export function Toasts() {
  const { toasts } = useApp()
  return (
    <div className="toasts">
      {toasts.map((toast) => (
        <div className="toast" key={toast.id} style={{ borderLeft: `3px solid ${toast.color}` }}>
          <span className="toast-dot" style={{ background: toast.color }} />
          <span className="toast-text">{toast.text}</span>
        </div>
      ))}
    </div>
  )
}

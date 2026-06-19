import { useApp } from '@/state/app'
import { useDeleteTag } from '@/hooks/queries'
import { ApiError } from '@/lib/api'
import { shortDigest } from '@/lib/format'
import { TrashIcon, AlertTriangleIcon } from '@/components/icons'

export function ConfirmDialog() {
  const { confirm, cancelDelete, t, addToast } = useApp()
  const del = useDeleteTag()

  if (!confirm) return null
  const { repo, tag, digest } = confirm

  const onConfirm = () => {
    del.mutate(
      { repo, tag },
      {
        onSuccess: () => {
          addToast(t.toast_deleted, 'var(--green)')
          cancelDelete()
        },
        onError: (e) => {
          const msg = e instanceof ApiError && e.status === 405 ? t.toast_405 : (e as Error).message
          addToast(msg, 'var(--red)')
          cancelDelete()
        },
      },
    )
  }

  return (
    <div className="overlay confirm" onClick={cancelDelete}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <div className="dialog-body">
          <div className="dialog-head">
            <div className="danger-badge">
              <TrashIcon size={20} />
            </div>
            <div style={{ minWidth: 0 }}>
              <div className="dialog-title">{t.confirm_title}</div>
              <div className="dialog-sub">{t.confirm_sub}</div>
            </div>
          </div>
          <div className="dialog-target">
            <div className="dialog-target-name">
              {repo}:{tag}
            </div>
            <div className="dialog-target-digest">{shortDigest(digest)}</div>
          </div>
          <div className="dialog-warn">
            <AlertTriangleIcon size={16} stroke="var(--amber)" style={{ flex: 'none', marginTop: 1 }} />
            <span>{t.confirm_warn}</span>
          </div>
        </div>
        <div className="dialog-actions">
          <button className="btn-ghost" onClick={cancelDelete} disabled={del.isPending}>
            {t.cancel}
          </button>
          <button className="btn-danger" onClick={onConfirm} disabled={del.isPending}>
            {del.isPending ? t.loading : t.delete_tag}
          </button>
        </div>
      </div>
    </div>
  )
}

// Typed client for the RegistryUI Go backend. Cookies carry the session, so
// every request is sent with credentials.

export interface Platform {
  os: string
  architecture: string
  variant?: string
  digest: string
  size: number
}

export interface Layer {
  digest: string
  size: number
  createdBy: string
}

export interface TagDetails {
  name: string
  digest: string
  mediaType: string
  size: number
  created?: string
  architecture?: string
  os?: string
  entrypoint?: string[]
  cmd?: string[]
  workingDir?: string
  env?: string[]
  labels?: Record<string, string>
  layers?: Layer[]
  isIndex: boolean
  platforms?: Platform[]
}

export interface RepoSummary {
  name: string
  tagCount: number
  size: number
  updated?: string
  description?: string
}

export interface Stats {
  repositories: number
  tags: number
  storage: number
}

export interface SessionInfo {
  registryUrl: string
  username: string
}

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    credentials: 'include',
    headers: { Accept: 'application/json', ...(init?.body ? { 'Content-Type': 'application/json' } : {}) },
    ...init,
  })
  if (!res.ok) {
    let message = res.statusText
    try {
      const body = await res.json()
      if (body?.error) message = body.error
    } catch {
      /* keep statusText */
    }
    throw new ApiError(message, res.status)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

const enc = encodeURIComponent

export const api = {
  // auth
  defaults: () => request<SessionInfo>('/api/defaults'),
  session: () => request<SessionInfo>('/api/session'),
  login: (registryUrl: string, username: string, password: string) =>
    request<SessionInfo>('/api/session', {
      method: 'POST',
      body: JSON.stringify({ registryUrl, username, password }),
    }),
  logout: () => request<void>('/api/session', { method: 'DELETE' }),

  // registry
  stats: () => request<Stats>('/api/stats'),
  listRepositories: () =>
    request<{ repositories: string[] }>('/api/repositories').then((r) => r.repositories),
  repoSummary: (repo: string) => request<RepoSummary>(`/api/repository?repo=${enc(repo)}`),
  listTags: (repo: string) =>
    request<{ repository: string; tags: string[] }>(`/api/tags?repo=${enc(repo)}`).then((r) => r.tags),
  tagDetails: (repo: string, tag: string) =>
    request<TagDetails>(`/api/tag?repo=${enc(repo)}&tag=${enc(tag)}`),
  deleteTag: (repo: string, tag: string) =>
    request<void>(`/api/tag?repo=${enc(repo)}&tag=${enc(tag)}`, { method: 'DELETE' }),
}

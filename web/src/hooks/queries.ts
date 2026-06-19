import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

export const keys = {
  health: ['health'] as const,
  stats: ['stats'] as const,
  repositories: ['repositories'] as const,
  repoSummary: (repo: string) => ['repoSummary', repo] as const,
  tags: (repo: string) => ['tags', repo] as const,
  tag: (repo: string, tag: string) => ['tag', repo, tag] as const,
}

export function useStats() {
  return useQuery({ queryKey: keys.stats, queryFn: api.stats, staleTime: 60_000 })
}

export function useRepositories() {
  return useQuery({ queryKey: keys.repositories, queryFn: api.listRepositories })
}

export function useRepoSummary(repo: string) {
  return useQuery({
    queryKey: keys.repoSummary(repo),
    queryFn: () => api.repoSummary(repo),
    staleTime: 60_000,
  })
}

export function useTags(repo: string) {
  return useQuery({
    queryKey: keys.tags(repo),
    queryFn: () => api.listTags(repo),
    enabled: !!repo,
  })
}

export function useTagDetails(repo: string, tag: string) {
  return useQuery({
    queryKey: keys.tag(repo, tag),
    queryFn: () => api.tagDetails(repo, tag),
    enabled: !!repo && !!tag,
  })
}

export function useDeleteTag() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ repo, tag }: { repo: string; tag: string }) => api.deleteTag(repo, tag),
    onSuccess: (_data, { repo }) => {
      qc.invalidateQueries({ queryKey: keys.tags(repo) })
      qc.invalidateQueries({ queryKey: keys.stats })
      qc.invalidateQueries({ queryKey: keys.repoSummary(repo) })
    },
  })
}

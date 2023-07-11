import { useMemo } from 'react'

import { HTTPResponse, map } from './useHTTPRequest'
import { useReposByOwner } from './useReposByOwner'

export type Repo = {
  owner: string
  name: string
  isInstalled: boolean
  environmentCount: number
  lastDeployedAt: string | null
}
export const useRepo = (
  owner: string,
  repo: string
): HTTPResponse<Repo | null> => {
  const res = useReposByOwner(owner)

  return useMemo(
    () => map(res, (repos) => repos.find((r) => r.name === repo) ?? null),
    [res, repo]
  )
}

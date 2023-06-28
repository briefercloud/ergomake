import { useMemo } from 'react'

import { useEnvironmentsByRepo } from './useEnvironmentsByRepo'
import { HTTPResponse, map } from './useHTTPRequest'

export type EnvironmentStatus =
  | 'pending'
  | 'building'
  | 'success'
  | 'degraded'
  | 'limited'
  | 'stale'

export type EnvironmentService = {
  id: string
  name: string
  url: string
  build: string
}

export type DegradedReason = {
  type: 'compose-not-found'
}

export type Environment = {
  id: string
  branch: string
  source: 'cli' | 'pr'
  status: EnvironmentStatus
  services: EnvironmentService[]
  createdAt: string
  degradedReason: DegradedReason | null
}

export const useEnvironment = (
  owner: string,
  repo: string,
  id: string
): HTTPResponse<Environment | null> => {
  const envs = useEnvironmentsByRepo(owner, repo)

  return useMemo(
    () => map(envs, (envs) => envs.find((e) => e.id === id) ?? null),
    [envs, id]
  )
}

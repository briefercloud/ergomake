import { EnvironmentStatus } from './useEnvironment'
import { UseHTTPRequest, useOptionalHTTPRequest } from './useHTTPRequest'

export type PublicEnvironment = {
  id: string
  owner: string
  repo: string
  status: EnvironmentStatus
  areServicesAlive: boolean
}

export const usePublicEnvironment = (
  id: string
): UseHTTPRequest<PublicEnvironment | null> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/environments/${id}/public`

  return useOptionalHTTPRequest(url)
}

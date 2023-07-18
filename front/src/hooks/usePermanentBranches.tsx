import { useMemo } from 'react'

import {
  HTTPResponse,
  andThen,
  isLoading,
  useHTTPMutation,
  useHTTPRequest,
} from './useHTTPRequest'

export type PermanentBranch = {
  name: string
}
type Payload = { branches: string[] }
type UsePermanentBranches = [
  HTTPResponse<PermanentBranch[]>,
  (payload: Payload) => void,
]

export const usePermanentBranches = (
  owner: string,
  repo: string
): UsePermanentBranches => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/owner/${owner}/repos/${repo}/permanent-branches`

  const [initial] = useHTTPRequest<PermanentBranch[]>(url)
  const [vars, update] = useHTTPMutation<Payload, PermanentBranch[]>(url)

  return useMemo((): UsePermanentBranches => {
    if (vars._tag === 'pristine') {
      return [initial, update]
    }

    if (isLoading(vars)) {
      return [andThen(initial, (i) => ({ ...i, refreshing: true })), update]
    }

    return [vars, update]
  }, [initial, vars, update])
}

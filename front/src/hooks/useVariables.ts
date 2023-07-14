import { useMemo } from 'react'

import {
  HTTPResponse,
  andThen,
  isLoading,
  useHTTPMutation,
  useHTTPRequest,
} from './useHTTPRequest'

export type Variable = {
  name: string
  value: string
  branch: string | null
}

type UseVariables = [HTTPResponse<Variable[]>, (variables: Variable[]) => void]

export const useVariables = (owner: string, repo: string): UseVariables => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/owner/${owner}/repos/${repo}/variables`

  const [initial] = useHTTPRequest<Variable[]>(url)
  const [vars, update] = useHTTPMutation<Variable[]>(url)

  return useMemo((): UseVariables => {
    if (vars._tag === 'pristine') {
      return [initial, update]
    }

    if (isLoading(vars)) {
      return [andThen(initial, (i) => ({ ...i, refreshing: true })), update]
    }

    return [vars, update]
  }, [initial, vars, update])
}

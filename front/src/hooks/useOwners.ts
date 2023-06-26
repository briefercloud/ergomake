import { sortBy } from 'ramda'
import { useMemo } from 'react'

import { HTTPResponse, map, useHTTPRequest } from './useHTTPRequest'

export type Owner = {
  login: string
  avatar: string
  isPaying: boolean
}
export const useOwners = (): HTTPResponse<Owner[]> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/github/user/organizations`
  const res = useHTTPRequest<Owner[]>(url)[0]

  return useMemo(
    () => map(res, (owners) => sortBy((o) => o.login, owners)),
    [res]
  )
}

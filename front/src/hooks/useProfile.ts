import { useMemo } from 'react'

import {
  HTTPResponse,
  isError,
  map,
  orElse,
  useHTTPRequest,
} from './useHTTPRequest'
import { Owner, useOwners } from './useOwners'

export type Profile = {
  username: string
  avatar: string
  owners: Owner[]
}
export const useProfile = (): HTTPResponse<Profile | null> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/auth/profile`

  const [res] = useHTTPRequest<Profile>(url)
  const owners = useOwners(true)

  return useMemo(() => {
    if (isError(res) && res.err._tag === 'authentication') {
      return { _tag: 'success', body: null, loading: false, refetching: false }
    }

    return map(res, (profile) => ({
      ...profile,
      owners: orElse(owners, []),
    }))
  }, [res, owners])
}

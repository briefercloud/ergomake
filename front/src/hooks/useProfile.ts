import { useMemo } from 'react'

import {
  HTTPResponse,
  andThen,
  isError,
  map,
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
  const owners = useOwners()

  return useMemo(() => {
    if (isError(res) && res.err._tag === 'authentication') {
      return { _tag: 'success', body: null, loading: false, refreshing: false }
    }

    return andThen(res, (profile) =>
      map(owners, (owners) => ({ ...profile.body, owners }))
    )
  }, [res, owners])
}

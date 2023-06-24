import { useMemo } from 'react'

import { HTTPResponse, map } from './useHTTPRequest'
import { Owner, useOwners } from './useOwners'

export const useOwner = (login: string): HTTPResponse<Owner | null> => {
  const res = useOwners()

  return useMemo(
    () =>
      map(
        res,
        (owners) => owners.find((owner) => owner.login === login) ?? null
      ),
    [res, login]
  )
}

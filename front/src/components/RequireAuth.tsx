import React from 'react'
import { Navigate, useLocation } from 'react-router-dom'

import { fold } from '../hooks/useHTTPRequest'
import { Profile, useProfile } from '../hooks/useProfile'
import ErrorLayout from '../layouts/Error'
import Loading from '../pages/Loading'

interface RequireAuthProps {
  children: (profile: Profile) => React.ReactNode
}
export function RequireAuth({ children }: RequireAuthProps) {
  const profile = useProfile()

  const location = useLocation()
  const originalUrl = `${window.location.protocol}//${window.location.host}${location.pathname}`

  return fold(profile, {
    onError: () => <ErrorLayout />,
    onLoading: () => <Loading />,
    onSuccess: ({ body }) =>
      body ? (
        <>{children(body)}</>
      ) : (
        <Navigate to={`/login?redirectUrl=${originalUrl}`} />
      ),
  })
}

interface RequireNoAuthProps {
  children: React.ReactNode
}
export function RequireNoAuth({ children }: RequireNoAuthProps) {
  const profile = useProfile()

  return fold(profile, {
    onError: () => <ErrorLayout />,
    onLoading: () => <Loading />,
    onSuccess: ({ body }) =>
      body ? (
        <Navigate to={`/gh/${body.owners[0]?.login ?? body.username}`} />
      ) : (
        <>{children}</>
      ),
  })
}

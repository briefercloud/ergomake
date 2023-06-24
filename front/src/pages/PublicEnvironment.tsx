import * as dfns from 'date-fns'
import React, { useEffect, useState } from 'react'
import { Navigate, useParams, useSearchParams } from 'react-router-dom'

import { isLoading, isSuccess } from '../hooks/useHTTPRequest'
import { usePublicEnvironment } from '../hooks/usePublicEnvironment'
import Layout from '../layouts/PublicEnvironment'

function PublicEnvironment() {
  const params = useParams()
  const [search] = useSearchParams()
  const redirect = search.get('redirect')
  const [environmentRes, refetch] = usePublicEnvironment(params.env ?? '', true)
  const [waitingPortsSince, setWaitingPortsSince] = useState(0)

  useEffect(() => {
    const loading =
      isLoading(environmentRes) ||
      (isSuccess(environmentRes) && environmentRes.loading)
    if (loading) {
      return
    }

    if (
      isSuccess(environmentRes) &&
      environmentRes.body?.status === 'success' &&
      environmentRes.body?.areServicesAlive
    ) {
      window.location.href = `${window.location.protocol}//${redirect}`
      return
    }

    const timeout = setTimeout(() => {
      refetch()
    }, 5000)

    return () => {
      clearTimeout(timeout)
    }
  }, [environmentRes, refetch, redirect])

  useEffect(() => {
    setWaitingPortsSince((t) => {
      if (t !== 0) {
        return t
      }

      if (
        isSuccess(environmentRes) &&
        environmentRes.body?.status === 'success'
      ) {
        return new Date().getTime()
      }

      return t
    })
  }, [environmentRes, setWaitingPortsSince])

  if (!params.env || !redirect) {
    return <Navigate to="/" />
  }

  const showLogsHint =
    waitingPortsSince !== 0 &&
    dfns.differenceInMinutes(new Date(), new Date(waitingPortsSince)) > 2

  return <Layout environment={environmentRes} showLogsHint={showLogsHint} />
}

export default PublicEnvironment

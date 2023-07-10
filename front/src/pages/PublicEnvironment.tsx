import ArrowPathIcon from '@heroicons/react/24/solid/ArrowPathIcon'
import Logo from '../components/Logo'
import * as dfns from 'date-fns'
import { useEffect, useState } from 'react'
import { Link, Navigate, useParams, useSearchParams } from 'react-router-dom'

import { isLoading, isSuccess } from '../hooks/useHTTPRequest'
import { usePublicEnvironment } from '../hooks/usePublicEnvironment'

function PublicEnvironment() {
  const params = useParams()
  const [search] = useSearchParams()
  const redirect = search.get('redirect')
  const [environmentRes, refetch] = usePublicEnvironment(params.env ?? '')
  const [waitingPortsSince, setWaitingPortsSince] = useState(0)

  useEffect(() => {
    const loading = isLoading(environmentRes)
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

  if (isSuccess(environmentRes) && !environmentRes.body) {
    return <Navigate to="/" />
  }

  let logsLink = ''
  let caption = 'Waking up services'
  if (isSuccess(environmentRes)) {
    if (!environmentRes.body) {
      caption = 'This environment is terminated.'
    } else if (environmentRes.body.status === 'success') {
      const env = environmentRes.body
      if (env.areServicesAlive) {
        caption = 'Redirecting'
      } else {
        caption = 'Waiting for services to start accepting connections'
        if (showLogsHint) {
          logsLink = `/gh/${env.owner}/repos/${env.repo}/envs/${env.id}`
        }
      }
    }
  }

  return (
    <div className="absolute inset-0 flex items-center justify-center bg-gray-900">
      <div className="flex items-center flex-col">
        <div className="flex items-center">
          <Logo />
          <ArrowPathIcon className="ml-4 animate-spin h-8 w-8 text-white" />
        </div>
        <p className="text-white pb-2">{caption}</p>
        {showLogsHint && (
          <p className="text-white">
            Taking too long? You might want to{' '}
            <Link
              className="text-primary-500 hover:underline hover:text-primary-400"
              to={logsLink}
            >
              check the logs
            </Link>
            .
          </p>
        )}
      </div>
    </div>
  )
}

export default PublicEnvironment

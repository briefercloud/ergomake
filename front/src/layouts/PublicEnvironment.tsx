import ArrowPathIcon from '@heroicons/react/24/solid/ArrowPathIcon'
import React from 'react'
import { Link, Navigate } from 'react-router-dom'

import Logo from '../components/Logo'
import { HTTPResponse, isSuccess } from '../hooks/useHTTPRequest'
import { PublicEnvironment } from '../hooks/usePublicEnvironment'

interface Props {
  environment: HTTPResponse<PublicEnvironment | null>
  showLogsHint: boolean
}
function Layout({ environment, showLogsHint }: Props) {
  if (isSuccess(environment) && !environment.body) {
    return <Navigate to="/" />
  }

  let logsLink = ''
  let caption = 'Waking up services'
  if (isSuccess(environment) && environment.body?.status === 'success') {
    const env = environment.body
    if (env.areServicesAlive) {
      caption = 'Redirecting'
    } else {
      caption = 'Waiting for services to start accepting connections'
      if (showLogsHint) {
        logsLink = `/gh/${env.owner}/repos/${env.repo}/envs/${env.id}`
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
export default Layout

import React, { ComponentProps, useCallback, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'

import useBuildLogs from '../hooks/useBuildLogs'
import { useEnvironment } from '../hooks/useEnvironment'
import { isError, isLoading } from '../hooks/useHTTPRequest'
import useLiveLogs from '../hooks/useLiveLogs'
import { useOwner } from '../hooks/useOwner'
import { Profile } from '../hooks/useProfile'
import { useRepo } from '../hooks/useRepo'
import Layout from '../layouts/Environment'
import Loading from './Loading'

interface Props {
  profile: Profile
}
function Environment(props: Props) {
  const [logsSwitch, setLogsSwitch] =
    useState<ComponentProps<typeof Layout>['logsSwitch']>('build')
  const params = useParams()
  const ownerRes = useOwner(params.owner ?? '')
  const repoRes = useRepo(params.owner ?? '', params.repo ?? '')
  const environmentRes = useEnvironment(
    params.owner ?? '',
    params.repo ?? '',
    params.env ?? ''
  )
  const onToggleLogsSwitch = useCallback(() => {
    setLogsSwitch((s) => (s === 'live' ? 'build' : 'live'))
  }, [setLogsSwitch])
  const [buildLogs, buildLogsErr, buildLogsRetry] = useBuildLogs(
    params.env ?? ''
  )
  const [liveLogs, liveLogsErr, liveLogsRetry] = useLiveLogs(params.env ?? '')
  const [currentService, setCurrentService] = useState(0)

  if (!params.owner || !params.repo || !params.env) {
    return <Navigate to="/" />
  }

  const loading =
    isLoading(ownerRes) || isLoading(repoRes) || isLoading(environmentRes)
  if (loading) {
    return <Loading />
  }

  const error = isError(ownerRes) || isError(repoRes) || isError(environmentRes)
  if (error) {
    // TODO: show error to user, etc...
    return <Navigate to="/" />
  }

  const owner = ownerRes.body
  const repo = repoRes.body
  const environment = environmentRes.body

  if (!owner || !repo || !environment) {
    return <Navigate to={`/gh/${params.owner}/repos/${params.repo}`} />
  }

  // TODO: handle logs error
  void buildLogsErr
  void buildLogsRetry
  void liveLogsErr
  void liveLogsRetry

  return (
    <Layout
      profile={props.profile}
      owner={owner}
      repo={repo}
      environment={environment}
      logsSwitch={logsSwitch}
      onToggleLogsSwitch={onToggleLogsSwitch}
      buildLogs={buildLogs}
      liveLogs={liveLogs}
      currentService={currentService}
      onChangeService={setCurrentService}
    />
  )
}

export default Environment

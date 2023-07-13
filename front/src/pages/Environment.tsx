import AnsiToHTML from 'ansi-to-html'
import classNames from 'classnames'
import * as dfns from 'date-fns'
import { useEffect, useMemo, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'

import Layout from '../components/Layout'
import Loading from '../components/Loading'
import Select from '../components/Select'
import useBuildLogs from '../hooks/useBuildLogs'
import { useEnvironment } from '../hooks/useEnvironment'
import { isError, isLoading, orElse } from '../hooks/useHTTPRequest'
import useLiveLogs from '../hooks/useLiveLogs'
import { useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { EnvironmentStatusStyle } from './Environments'

const secondaryNavigation: Array<{ name: string; logType: LogType }> = [
  { name: 'Live logs', logType: 'live' },
  { name: 'Build logs', logType: 'build' },
]

const stats = [
  { name: 'Number of deploys', value: '405' },
  { name: 'Average deploy time', value: '3.65', unit: 'mins' },
  { name: 'Number of servers', value: '3' },
  { name: 'Success rate', value: '98.5%' },
]

type Props = {
  profile: Profile
}

type LogType = 'build' | 'live'

const converter = new AnsiToHTML()

function Environment({ profile }: Props) {
  const params = useParams()
  const ownersRes = useOwners()
  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])
  const owner = useMemo(
    () =>
      owners.find((org) => org.login === params.owner) ?? {
        login: params.owner ?? '',
        avatar: '',
        isPaying: false,
      },
    [owners, params.owner]
  )

  const [environmentRes, refetchEnv] = useEnvironment(
    params.owner ?? '',
    params.repo ?? '',
    params.env ?? ''
  )
  useEffect(() => {
    const interval = setInterval(refetchEnv, 5000)
    return () => {
      clearInterval(interval)
    }
  }, [refetchEnv])

  const [logType, setLogsType] = useState<LogType>('build')

  const [buildLogs, buildLogsErr, buildLogsRetry] = useBuildLogs(
    params.env ?? ''
  )
  void buildLogsErr
  void buildLogsRetry

  const [liveLogs, liveLogsErr, liveLogsRetry] = useLiveLogs(params.env ?? '')
  void liveLogsErr
  void liveLogsRetry

  const [currentServiceIndex, setCurrentServiceIndex] = useState(0)

  const loading = isLoading(ownersRes) || isLoading(environmentRes)
  if (loading) {
    return <Loading />
  }

  if (isError(ownersRes) || isError(environmentRes)) {
    // TODO: proper error
    return <Navigate to="/" />
  }

  const environment = environmentRes.body
  if (!owner || !environment) {
    return <Navigate to={`/gh/${params.owner}/repos/${params.repo}`} />
  }

  const currentService = environment.services[currentServiceIndex]
  if (!currentService) {
    // TODO: proper error
    return <Navigate to="/" />
  }

  const selectOptions = environment.services.map((service, i) => ({
    label: service.name,
    value: i,
  }))

  const logs = logType === 'build' ? buildLogs : liveLogs

  const selectedLogs = (logs[currentService.id] ?? []).map((log, i) => {
    const html = converter.toHtml(log.message)
    return (
      <pre
        key={i}
        className="text-white whitespace-pre-wrap  last:mb-16"
        dangerouslySetInnerHTML={{ __html: html }}
      />
    )
  })

  const pages = [
    {
      name: 'Repositories',
      href: `/gh/${owner.login}`,
      label: 'Repositories',
    },
    {
      name: params.repo ?? '',
      href: `/gh/${owner.login}/repos/${params.repo}`,
      label: params.repo ?? '',
    },
    {
      name: environment.branch,
      href: `/gh/${owner.login}/repos/${params.repo}/envs/${params.env}`,
      label: environment.branch,
    },
  ]

  return (
    <Layout profile={profile} pages={pages}>
      <header>
        {/* Heading */}
        <div className="flex flex-col items-start justify-between gap-x-8 gap-y-4 bg-white dark:bg-neutral-950 px-4 py-4 sm:flex-row sm:items-center sm:px-6 lg:px-8">
          <div>
            <div className="flex items-center gap-x-3">
              <div
                className={classNames(
                  'flex-none rounded-full p-1',
                  EnvironmentStatusStyle[environment.status]
                )}
              >
                <div className="h-2 w-2 rounded-full bg-current" />
              </div>
              <h1 className="flex gap-x-3 text-base leading-7">
                <span className="font-semibold text-gray-800 dark:text-neutral-300">
                  {params.repo}
                </span>
                <span className="text-gray-600 dark:text-neutral-500">/</span>
                <span className="font-semibold text-gray-800 dark:text-neutral-300">
                  {environment.branch}
                </span>
              </h1>
            </div>
            <p className="mt-2 text-xs leading-6 text-gray-400">
              Last deployed at{' '}
              {dfns.formatRelative(new Date(environment.createdAt), new Date())}
            </p>
          </div>
          {/*
          <div className="order-first flex-none rounded-full bg-primary-400/10 px-2 py-1 text-xs font-medium text-primary-400 ring-1 ring-inset ring-primary-400/30 sm:order-none">
            Production
          </div>
          */}
        </div>

        {/* Stats */}
        <div className="hidden grid grid-cols-1 bg-white sm:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat, statIdx) => (
            <div
              key={stat.name}
              className={classNames(
                statIdx % 2 === 1
                  ? 'sm:border-l'
                  : statIdx === 2
                  ? 'lg:border-l'
                  : '',
                'border-t border-gray-200 dark:border-neutral-800 py-6 px-4 sm:px-6 lg:px-8'
              )}
            >
              <p className="text-sm font-medium leading-6 text-gray-400">
                {stat.name}
              </p>
              <p className="mt-2 flex items-baseline gap-x-2">
                <span className="text-4xl font-semibold tracking-tight text-gray-800">
                  {stat.value}
                </span>
                {stat.unit ? (
                  <span className="text-sm text-gray-400">{stat.unit}</span>
                ) : null}
              </p>
            </div>
          ))}
        </div>
      </header>

      {/* Secondary navigation */}
      <nav className="flex border-y border-gray-200 dark:border-neutral-800 border-b-0">
        <ul className="flex justify-between min-w-full flex-none text-sm font-semibold leading-6 text-gray-800 [&>li]:h-full [&>li]:flex [&>li]:grow [&>li]:items-center  [&>li]:justify-center items-center [&>li]:flex-1">
          <li className="bg-red">
            <Select options={selectOptions} onChange={setCurrentServiceIndex} />
          </li>
          {secondaryNavigation.map((item) => (
            <li
              key={item.name}
              className={classNames(
                item.logType === logType
                  ? 'text-primary-400 dark:text-primary-200 shadow-inner shadow-gray-900 bg-gray-100 dark:bg-neutral-800'
                  : 'text-gray-400',
                'hover:bg-gray-100 dark:hover:bg-neutral-800 hover:cursor-pointer'
              )}
              onClick={() => setLogsType(item.logType)}
            >
              {item.name}
            </li>
          ))}
        </ul>
      </nav>

      <div className="bg-gray-800 dark:bg-neutral-700 font-mono p-4 overflow-y-scroll overflow-x-scroll h-full">
        {selectedLogs}
      </div>
    </Layout>
  )
}

export default Environment

import { CubeIcon } from '@heroicons/react/24/outline'
import classNames from 'classnames'
import * as dfns from 'date-fns'
import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'

import EmptyState from '../components/EmptyState'
import Layout from '../components/Layout'
import List from '../components/List'
import PermanentBranchesInput from '../components/PermanentBranchesInput'
import VariablesInput from '../components/VariablesInput'
import { EnvironmentStatus } from '../hooks/useEnvironment'
import { useEnvironmentsByRepo } from '../hooks/useEnvironmentsByRepo'
import { isSuccess, orElse } from '../hooks/useHTTPRequest'
import { useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'

interface Props {
  profile: Profile
}

export const EnvironmentStatusStyle: Record<EnvironmentStatus, string> = {
  pending: 'text-rose-400 bg-rose-400/30',
  building: 'text-blue-400 bg-blue-400/30',
  success: 'text-green-400 bg-blue-400/20',
  degraded: 'text-red-400 bg-red-400/20',
  limited: 'text-yellow-400 bg-yellow-400/20',
  stale: 'text-gray-500 bg-gray-100/30',
}

const EnvironmentStatusText: Record<EnvironmentStatus, string> = {
  pending: 'Pending',
  building: 'Building',
  success: 'Running',
  degraded: 'Failed',
  limited: 'Above limits',
  stale: 'Sleeping',
}

type TabName = 'branches' | 'envVars' | 'permanentBranches'

const secondaryNavigation: Array<{ name: string; tabName: TabName }> = [
  { name: 'Branches', tabName: 'branches' },
  { name: 'Environment Variables', tabName: 'envVars' },
  { name: 'Permanent Branches', tabName: 'permanentBranches' },
]

const StatusBall = ({ status }: { status: EnvironmentStatus }) => {
  return (
    <div
      className={classNames(
        EnvironmentStatusStyle[status],
        'flex-none rounded-full p-1'
      )}
    >
      <div className="h-2 w-2 rounded-full bg-current" />
    </div>
  )
}

const Environments = ({ profile }: Props) => {
  const params = useParams()

  const [currentTab, setCurrentTab] = useState<TabName>('branches')
  const ownersRes = useOwners()
  const owner = useMemo(
    () =>
      orElse(ownersRes, []).find((org) => org.login === params.owner) ?? {
        login: params.owner ?? '',
        avatar: '',
        isPaying: false,
      },
    [ownersRes, params.owner]
  )

  const [envsRes, refetchEnvs] = useEnvironmentsByRepo(
    owner.login,
    params.repo ?? ''
  )
  useEffect(() => {
    if (currentTab !== 'branches') {
      return
    }

    refetchEnvs()
    const interval = setInterval(refetchEnvs, 5000)
    return () => {
      clearInterval(interval)
    }
  }, [currentTab, refetchEnvs])
  const envs = useMemo(
    () =>
      orElse(envsRes, []).sort((a, b) =>
        b.createdAt.localeCompare(a.createdAt)
      ),
    [envsRes]
  )

  const envItems = useMemo(
    () =>
      envs.map((env) => ({
        name: env.branch,
        statusBall: <StatusBall status={env.status} />,
        descriptionLeft: EnvironmentStatusText[env.status],
        descriptionRight: `Created ${dfns.formatRelative(
          new Date(env.createdAt),
          new Date()
        )}`,
        url: `/gh/${owner.login}/repos/${params.repo}/envs/${env.id}`,
        data: env,
      })),
    [envs, owner.login, params.repo]
  )

  const pages = useMemo(
    () => [
      {
        name: 'Repositories',
        href: `/gh/${owner.login}`,
        label: 'Projects',
      },
      {
        name: params.repo ?? '',
        href: `/gh/${owner.login}/repos/${params.repo}`,
        label: params.repo ?? '',
      },
    ],
    [owner.login, params.repo]
  )

  const emptyStateComponent =
    isSuccess(envsRes) && !envsRes.refreshing ? (
      <EmptyState
        icon={CubeIcon}
        title="No environments available"
        description="Configure a permanent branch or create a pull-request to deploy an environment."
      />
    ) : null

  return (
    <Layout profile={profile} pages={pages}>
      <div className="bg-white dark:bg-neutral-950 border-b border-gray-200 dark:border-neutral-800">
        <div className="flex flex-col items-start justify-between gap-x-8 gap-y-4 px-4 py-4 sm:flex-row sm:items-center sm:px-6 lg:px-8">
          <div className="flex items-center gap-x-3 h-20">
            <h1 className="flex text-2xl font-bold tracking-tight sm:text-4xl">
              <span className="font-semibold text-gray-800 dark:text-gray-200">
                {params.repo}
              </span>
            </h1>
          </div>
        </div>

        <nav className="flex">
          <ul className="flex min-w-full flex-none gap-x-6 px-4 text-sm font-semibold leading-6 text-gray-400 sm:px-6 lg:px-8">
            {secondaryNavigation.map((item) => (
              <li
                key={item.name}
                className={classNames(
                  item.tabName === currentTab
                    ? 'text-primary-400 border-b-2 border-primary-400'
                    : '',
                  'pb-2 px-2 hover:text-primary-300 hover:border-primary-300 hover:cursor-pointer hover:border-primary-300'
                )}
                onClick={() => setCurrentTab(item.tabName)}
              >
                {item.name}
              </li>
            ))}
          </ul>
        </nav>
      </div>

      {currentTab === 'branches' && (
        <List items={envItems} emptyState={emptyStateComponent} />
      )}

      {currentTab === 'envVars' && (
        <VariablesInput owner={owner.login} repo={params.repo ?? ''} />
      )}

      {currentTab === 'permanentBranches' && (
        <PermanentBranchesInput owner={owner.login} repo={params.repo ?? ''} />
      )}
    </Layout>
  )
}

export default Environments

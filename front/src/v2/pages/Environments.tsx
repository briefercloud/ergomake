import * as dfns from 'date-fns'
import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'

import { EnvironmentStatus } from '../../hooks/useEnvironment'
import { useEnvironmentsByRepo } from '../../hooks/useEnvironmentsByRepo'
import { orElse } from '../../hooks/useHTTPRequest'
import { useOwners } from '../../hooks/useOwners'
import { Profile } from '../../hooks/useProfile'
import Layout from '../components/Layout'
import List from '../components/List'
import { classNames } from '../utils'

interface Props {
  profile: Profile
}

const EnvironmentStatusStyle: Record<EnvironmentStatus, string> = {
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

  const envsRes = useEnvironmentsByRepo(owner.login, params.repo ?? '')
  const [search, setSearch] = useState('')
  const envs = useMemo(
    () =>
      orElse(envsRes, [])
        .filter((e) => search === '' || e.branch.includes(search))
        .sort((a, b) => b.createdAt.localeCompare(a.createdAt)),
    [envsRes, search]
  )

  const envItems = envs.map((env) => ({
    name: env.branch,
    statusBall: <StatusBall status={env.status} />,
    descriptionLeft: EnvironmentStatusText[env.status],
    descriptionRight: `Created ${dfns.formatRelative(
      new Date(env.createdAt),
      new Date()
    )}`,
    url: `/v2/gh/${owner.login}/${params.repo}/envs/${env.branch}`,
  }))

  const pages = [
    { name: 'Repositories', href: `/v2/gh/${owner.login}`, label: 'Projects' },
    {
      name: params.repo ?? '',
      href: `/v2/gh/${owner.login}`,
      label: 'Projects',
    },
  ]

  return (
    <Layout profile={profile} pages={pages}>
      <List items={envItems} />
    </Layout>
  )
}

export default Environments

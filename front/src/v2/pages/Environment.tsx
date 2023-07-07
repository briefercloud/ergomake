import { MagnifyingGlassIcon } from '@heroicons/react/20/solid'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { orElse } from '../../hooks/useHTTPRequest'
import { useOwners } from '../../hooks/useOwners'
import { Profile } from '../../hooks/useProfile'
import Layout from '../components/Layout'
import Select from '../components/Select'
import { classNames } from '../utils'

const secondaryNavigation = [
  { name: 'Live logs', href: '#', current: true },
  { name: 'Build logs', href: '#', current: false },
]

const stats = [
  { name: 'Number of deploys', value: '405' },
  { name: 'Average deploy time', value: '3.65', unit: 'mins' },
  { name: 'Number of servers', value: '3' },
  { name: 'Success rate', value: '98.5%' },
]

const statuses = {
  Completed: 'text-green-400 bg-green-400/10',
  Error: 'text-rose-400 bg-rose-400/10',
}

const activityItems = [
  {
    user: {
      name: 'Michael Foster',
      imageUrl:
        'https://images.unsplash.com/photo-1519244703995-f4e0f30006d5?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80',
    },
    commit: '2d89f0c8',
    branch: 'main',
    status: 'Completed',
    duration: '25s',
    date: '45 minutes ago',
    dateTime: '2023-01-23T11:00',
  },
  // More items...
]

type Props = {
  profile: Profile
}

const Details = ({ profile }: Props) => {
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

  const pages = [
    {
      name: 'Repositories',
      href: `/v2/gh/${owner.login}`,
      label: 'Repositories',
    },
    {
      name: params.repo ?? '',
      href: `/v2/gh/${owner.login}/repos/${params.repo}`,
      label: params.repo ?? '',
    },
    {
      name: params.env ?? '',
      href: `/v2/gh/${owner.login}/repos/${params.repo}/envs/${params.env}`,
      label: params.env ?? '',
    },
  ]

  return (
    <Layout profile={profile} pages={pages}>
      <main>
        <header>
          {/* Heading */}
          <div className="flex flex-col items-start justify-between gap-x-8 gap-y-4 bg-white px-4 py-4 sm:flex-row sm:items-center sm:px-6 lg:px-8">
            <div>
              <div className="flex items-center gap-x-3">
                <div className="flex-none rounded-full bg-green-400/10 p-1 text-green-400">
                  <div className="h-2 w-2 rounded-full bg-current" />
                </div>
                <h1 className="flex gap-x-3 text-base leading-7">
                  <span className="font-semibold text-gray-800">
                    Planetaria
                  </span>
                  <span className="text-gray-600">/</span>
                  <span className="font-semibold text-gray-800">
                    mobile-api
                  </span>
                </h1>
              </div>
              <p className="mt-2 text-xs leading-6 text-gray-400">
                Deploys from GitHub via main branch
              </p>
            </div>
            <div className="order-first flex-none rounded-full bg-primary-400/10 px-2 py-1 text-xs font-medium text-primary-400 ring-1 ring-inset ring-primary-400/30 sm:order-none">
              Production
            </div>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 bg-white sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat, statIdx) => (
              <div
                key={stat.name}
                className={classNames(
                  statIdx % 2 === 1
                    ? 'sm:border-l'
                    : statIdx === 2
                    ? 'lg:border-l'
                    : '',
                  'border-t border-gray-200 py-6 px-4 sm:px-6 lg:px-8'
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
        <nav className="flex border-y border-gray-200">
          <ul className="flex justify-between min-w-full flex-none text-sm font-semibold leading-6 text-gray-800 [&>li]:h-full [&>li]:flex [&>li]:grow [&>li]:items-center  [&>li]:justify-center items-center">
            <li className="bg-red">
              <Select options={[{ label: 'API', value: 0 }]} />
            </li>
            {secondaryNavigation.map((item) => (
              <li
                key={item.name}
                className={classNames(
                  item.current ? 'text-primary-400' : 'text-gray-400',
                  'hover:bg-gray-100'
                )}
              >
                <a
                  className="w-full h-full flex justify-center items-center"
                  href={item.href}
                >
                  {item.name}
                </a>
              </li>
            ))}
          </ul>
        </nav>

        <div className="font-mono p-4 overflow-y-scroll">
          <p>Test</p>
        </div>
      </main>
    </Layout>
  )
}

export default Details

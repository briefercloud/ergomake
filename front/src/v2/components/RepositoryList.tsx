import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { Link } from 'react-router-dom'

import { Repo } from '../../hooks/useRepo'
import { classNames } from '../utils'

type Deployment = {
  id: number
  href: string
  projectName: string
  teamName: string
  statusText: string
  description: string
}

const deployments: Deployment[] = [
  {
    id: 1,
    href: '#',
    projectName: 'ios-app',
    teamName: 'Planetaria',
    statusText: 'Initiated 1m 32s ago',
    description: 'Deploys from GitHub',
  },
  {
    id: 2,
    href: '#',
    projectName: 'mobile-api',
    teamName: 'Planetaria',
    statusText: 'Deployed 3m ago',
    description: 'Deploys from GitHub',
  },
  {
    id: 3,
    href: '#',
    projectName: 'tailwindcss.com',
    teamName: 'Tailwind Labs',
    statusText: 'Deployed 3h ago',
    description: 'Deploys from GitHub',
  },
  {
    id: 4,
    href: '#',
    projectName: 'api.protocol.chat',
    teamName: 'Protocol',
    statusText: 'Failed to deploy 6d ago',
    description: 'Deploys from GitHub',
  },
]

type RepositoryListProps = {
  repos: Repo[]
}

const RepositoryList = ({ repos }: RepositoryListProps) => {
  return (
    <ul className="-mx-4 sm:-mx-6 lg:-mx-8">
      {repos.map((repo) => (
        <Link to={`/v2/gh/${repo.owner}/repos/${repo.name}/`}>
          <li
            key={repo.name}
            className="relative flex items-center space-x-4 py-4 px-8 hover:bg-gray-100 border-b border-gray-200/70 hover:cursor-pointer"
          >
            <div className="min-w-0 flex-auto space-y-2">
              <div className="flex items-center gap-x-3">
                <h2 className="min-w-0 text-lg font-semibold leading-6 text-primary-900">
                  <span className="whitespace-nowrap">{repo.name}</span>
                </h2>
              </div>
              <div className="mt-3 flex items-center gap-x-2.5 text-xs leading-5 text-gray-400">
                <p className="truncate">
                  {repo.environmentCount === 1
                    ? `${repo.environmentCount} environment`
                    : `${repo.environmentCount} environments`}{' '}
                </p>
                <svg
                  viewBox="0 0 2 2"
                  className="h-1 w-1 flex-none fill-gray-300"
                >
                  <circle cx={1} cy={1} r={1} />
                </svg>
                <p className="whitespace-nowrap">Deploys from GitHub</p>
              </div>
            </div>
            <ChevronRightIcon
              className="h-5 w-5 flex-none text-gray-400"
              aria-hidden="true"
            />
          </li>
        </Link>
      ))}
    </ul>
  )
}

export default RepositoryList

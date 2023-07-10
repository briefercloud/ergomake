import { PlusIcon, UserPlusIcon } from '@heroicons/react/20/solid'
import { useMemo } from 'react'
import { Navigate } from 'react-router-dom'

import Layout, { installationUrl } from '../components/Layout'
import { orElse } from '../hooks/useHTTPRequest'
import { useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'

interface NoInstallationProps {
  profile: Profile
}

const NoInstallation = ({ profile }: NoInstallationProps) => {
  const ownersRes = useOwners()
  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])

  if (owners.length > 0 && owners[0] !== undefined) {
    return <Navigate to={`/gh/${owners[0].login}`} />
  }

  return (
    <Layout profile={profile} pages={[]}>
      <div className="px-4 sm:px-6 lg:px-8 pt-6">
        <div className="relative block w-full rounded-lg border-2 border-dashed border-gray-300 p-12 text-center focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
          <div className="p-4">
            <UserPlusIcon className="mx-auto h-12 w-12 text-gray-400" />
          </div>
          <h3 className="mt-2 block text-sm font-semibold text-gray-900">
            No organizations available.
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            Allow access to an organization to get started.
          </p>

          <div className="pt-6">
            <a
              href={installationUrl}
              className="inline-flex items-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
            >
              <PlusIcon className="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
              Add organization
            </a>
          </div>
        </div>
      </div>
    </Layout>
  )
}

export default NoInstallation

import { PlusIcon } from '@heroicons/react/20/solid'
import { useCallback, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import ConfigureRepoModal from '../components/ConfigureRepoModal'
import Layout, { installationUrl } from '../components/Layout'
import RepositoryList from '../components/RepositoryList'
import { orElse } from '../hooks/useHTTPRequest'
import { useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Repo } from '../hooks/useRepo'
import { useReposByOwner } from '../hooks/useReposByOwner'

interface Props {
  profile: Profile
}

const Projects = ({ profile }: Props) => {
  const [configuring, setConfiguring] = useState<Repo | null>(null)
  const [configured, setConfigured] = useState<Set<string>>(new Set())
  const ownersRes = useOwners()
  const params = useParams<{ owner: string }>()

  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])

  const owner =
    owners.find((o) => o.login === params.owner)?.login ?? profile.username

  const pages = [
    { name: 'Repositories', href: `/gh/${owner}`, label: 'Projects' },
  ]

  const reposRes = useReposByOwner(params.owner ?? profile.username)
  const repos = useMemo(() => {
    const repos = orElse(reposRes, []).filter((r) => r.isInstalled)

    return repos
      .sort((a, b) => a.name.localeCompare(b.name))
      .map((r) => ({
        ...r,
        lastDeployedAt:
          r.lastDeployedAt ??
          (configured.has(r.name) ? new Date().toISOString() : null),
      }))
  }, [reposRes, configured])

  const onCloseConfiguring = useCallback(
    (success: boolean) => {
      if (configuring && success) {
        const repo = configuring.name
        setConfigured((c) => new Set([...Array.from(c), repo]))
      }
      setConfiguring(null)
    },
    [configuring, setConfiguring, setConfigured]
  )

  return (
    <Layout profile={profile} pages={pages}>
      <ConfigureRepoModal repo={configuring} onClose={onCloseConfiguring} />
      <div className="bg-white border-b border-gray-200 flex flex-col items-start justify-between gap-x-8 gap-y-4 bg-white px-4 py-4 sm:flex-row sm:items-center sm:px-6 lg:px-8  ">
        <h1 className="flex text-2xl tracking-tight font-semibold text-gray-800 sm:text-4xl h-20 items-center">
          Repositories
        </h1>

        <a
          href={installationUrl}
          className="inline-flex items-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          <PlusIcon className="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
          Add repository
        </a>
      </div>

      <RepositoryList repos={repos} onConfigure={setConfiguring} />
    </Layout>
  )
}

export default Projects

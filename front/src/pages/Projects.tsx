import { useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'

import Layout from '../components/Layout'
import RepositoryList from '../components/RepositoryList'
import { orElse } from '../hooks/useHTTPRequest'
import { useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { useReposByOwner } from '../hooks/useReposByOwner'

interface Props {
  profile: Profile
}

const Projects = ({ profile }: Props) => {
  const ownersRes = useOwners()
  const params = useParams<{ owner: string }>()

  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])

  const owner =
    owners.find((o) => o.login === params.owner)?.login ?? profile.username

  const pages = [
    { name: 'Repositories', href: `/gh/${owner}`, label: 'Projects' },
  ]

  const reposRes = useReposByOwner(params.owner ?? profile.username)
  const [search, setSearch] = useState('')
  const [repos, hasProjects] = useMemo(() => {
    const repos = orElse(reposRes, []).filter((r) => r.isInstalled)

    return [
      repos
        .sort((a, b) => a.name.localeCompare(b.name))
        .filter((r) => search === '' || r.name.includes(search)),
      repos.length > 0,
    ]
  }, [reposRes, search])

  return (
    <Layout profile={profile} pages={pages}>
      <RepositoryList repos={repos} />
    </Layout>
  )
}

export default Projects

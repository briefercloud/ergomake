import { useCallback, useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'

import { isLoading, orElse } from '../hooks/useHTTPRequest'
import { Owner, useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { useReposByOwner } from '../hooks/useReposByOwner'
import Layout from '../layouts/Projects'
import Loading from './Loading'

interface Props {
  profile: Profile
}

function Projects({ profile }: Props) {
  const ownersRes = useOwners()
  const params = useParams<{ owner: string }>()

  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])

  const owner = owners.find((o) => o.login === params.owner)

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

  const navigate = useNavigate()
  const onChangeOwner = useCallback(
    (owner: Owner) => {
      navigate(`/gh/${owner.login}`)
    },
    [navigate]
  )

  if (!owner && isLoading(ownersRes)) {
    return <Loading />
  }

  if (!owner) {
    return <Navigate to="/" />
  }

  return (
    <Layout
      profile={profile}
      owners={owners}
      owner={owner}
      onChangeOwner={onChangeOwner}
      loadingRepos={isLoading(reposRes)}
      repositories={repos}
      hasProjects={hasProjects}
      search={search}
      onChangeSearch={setSearch}
    />
  )
}

export default Projects

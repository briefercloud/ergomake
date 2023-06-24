import React, { useCallback, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { isLoading, isSuccess, orElse } from '../hooks/useHTTPRequest'
import { Owner, useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { useReposByOwner } from '../hooks/useReposByOwner'
import Layout from '../layouts/Projects'

interface Props {
  profile: Profile
}
function Projects({ profile }: Props) {
  const ownersRes = useOwners()

  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])

  const params = useParams()
  const owner: Owner = owners.find((o) => o.login === params.owner) ?? {
    login: params.owner ?? profile.username,
    avatar: profile.avatar,
    isPaying: false,
  }

  const reposRes = useReposByOwner(owner.login)
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

  return (
    <Layout
      profile={profile}
      owners={owners}
      owner={owner}
      onChangeOwner={onChangeOwner}
      loadingRepos={
        isLoading(reposRes) || (isSuccess(reposRes) && reposRes.loading)
      }
      repositories={repos}
      hasProjects={hasProjects}
      search={search}
      onChangeSearch={setSearch}
    />
  )
}

export default Projects

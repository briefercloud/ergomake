import React, { useCallback, useMemo, useState } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'

import useBool from '../hooks/useBool'
import { useEnvironmentsByRepo } from '../hooks/useEnvironmentsByRepo'
import { isLoading, isSuccess, orElse } from '../hooks/useHTTPRequest'
import { useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Repo } from '../hooks/useRepo'
import { useReposByOwner } from '../hooks/useReposByOwner'
import { useVariables } from '../hooks/useVariables'
import Layout from '../layouts/Environments'
import Loading from './Loading'

interface Props {
  profile: Profile
}
function Environments(props: Props) {
  const params = useParams()
  const navigate = useNavigate()

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

  const reposRes = useReposByOwner(owner.login)
  const repos = useMemo(() => {
    return orElse(reposRes, []).filter((repo) => repo.isInstalled)
  }, [reposRes])

  const repo: Repo = useMemo(
    () =>
      repos.find((repo) => repo.name === params.repo) ?? {
        owner: params.owner ?? '',
        name: params.repo ?? '',
        isInstalled: false,
        environmentCount: 0,
        variables: [],
      },
    [repos, params.owner, params.repo]
  )

  const onChangeRepo = useCallback(
    (repo: Repo) => {
      navigate(`/gh/${params.owner}/repos/${repo.name}`)
    },
    [navigate, params]
  )

  const envsRes = useEnvironmentsByRepo(owner.login, repo.name)
  const [search, setSearch] = useState('')
  const envs = useMemo(
    () =>
      orElse(envsRes, [])
        .filter((e) => search === '' || e.branch.includes(search))
        .sort((a, b) => b.createdAt.localeCompare(a.createdAt)),
    [envsRes, search]
  )

  const [variablesRes, onUpdateVariables] = useVariables(owner.login, repo.name)
  const updatingVariables =
    isLoading(variablesRes) || (isSuccess(variablesRes) && variablesRes.loading)
  const variables = useMemo(() => orElse(variablesRes, []), [variablesRes])
  const [
    managingVariables,
    { setTrue: onManageVariables, setFalse: onDiscardVariableManagement },
  ] = useBool(false)

  if (!params.owner || !params.repo) {
    return <Navigate to="/" />
  }

  const loading = ownersRes._tag === 'loading' || reposRes._tag === 'loading'
  if (loading) {
    return <Loading />
  }

  return (
    <Layout
      profile={props.profile}
      owner={owner}
      repo={repo}
      repos={repos}
      onChangeRepo={onChangeRepo}
      loadingEnvironments={
        isLoading(envsRes) || (isSuccess(envsRes) && envsRes.loading)
      }
      environments={envs}
      search={search}
      onChangeSearch={setSearch}
      variables={variables}
      onManageVariables={onManageVariables}
      managingVariables={managingVariables}
      onDiscardVariableManagement={onDiscardVariableManagement}
      onUpdateVariables={onUpdateVariables}
      updatingVariables={updatingVariables}
    />
  )
}

export default Environments

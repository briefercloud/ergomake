import React, { useCallback, useMemo } from 'react'
import { Navigate, useNavigate, useParams } from 'react-router-dom'

import useBool from '../hooks/useBool'
import { isError, isLoading, isSuccess, orElse } from '../hooks/useHTTPRequest'
import { Owner, useOwners } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Registry, useRegistries } from '../hooks/useRegistries'
import ErrorLayout from '../layouts/Error'
import Layout from '../layouts/Registries'
import Loading from './Loading'

interface Props {
  profile: Profile
}
function Environments({ profile }: Props) {
  const params = useParams()

  const ownersRes = useOwners()
  const owners = useMemo(() => orElse(ownersRes, []), [ownersRes])
  const owner = useMemo(
    () => owners.find((org) => org.login === params.owner),
    [owners, params.owner]
  )

  const navigate = useNavigate()
  const onChangeOwner = useCallback(
    (owner: Owner) => {
      navigate(`/gh/${owner.login}/registries`)
    },
    [navigate]
  )

  const [registriesRes, refetchRegistries] = useRegistries(
    params.owner ?? profile.username
  )
  const registries = orElse(registriesRes, [])
  const loadingRegistries = isLoading(registriesRes)

  const [adding, { setTrue: onOpenAdding, setFalse: setAddingToFalse }] =
    useBool(false)

  const onCloseAdding = useCallback(() => {
    setAddingToFalse()
    refetchRegistries()
  }, [setAddingToFalse, refetchRegistries])

  const onDelete = useCallback(
    (reg: Registry) => {
      const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/owner/${
        params.owner ?? profile.username
      }/registries/${reg.id}`

      fetch(url, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      }).then(() => {
        refetchRegistries()
      })
    },
    [profile, params.owner, refetchRegistries]
  )

  if (!params.owner) {
    return <Navigate to="/" />
  }

  const loading = isLoading(ownersRes)
  if (!owner && loading) {
    return <Loading />
  }

  if (!owner) {
    if (isError(ownersRes)) {
      return <ErrorLayout />
    }

    return <Navigate to="/" />
  }

  return (
    <Layout
      profile={profile}
      owner={owner}
      owners={owners}
      onChangeOwner={onChangeOwner}
      registries={registries}
      loading={loadingRegistries}
      adding={adding}
      onOpenAdding={onOpenAdding}
      onCloseAdding={onCloseAdding}
      onDelete={onDelete}
    />
  )
}

export default Environments

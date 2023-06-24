import React from 'react'
import { Navigate, useParams } from 'react-router-dom'

import { Profile } from '../hooks/useProfile'
import Layout from '../layouts/Purchase'

interface Props {
  profile: Profile
}
function Environments(props: Props) {
  const params = useParams()

  if (!params.owner) {
    return <Navigate to="/" />
  }

  return <Layout profile={props.profile} owner={params.owner} />
}

export default Environments

import React from 'react'

import { Profile } from '../hooks/useProfile'
import Layout from '../layouts/NoInstallation'

interface Props {
  profile: Profile
}
function NoInstallation(props: Props) {
  return <Layout profile={props.profile} />
}

export default NoInstallation

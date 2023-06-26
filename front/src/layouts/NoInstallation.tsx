import React from 'react'

import Button from '../components/Button'
import Navbar from '../components/Navbar'
import { Profile } from '../hooks/useProfile'
import { installationUrl } from './Projects'

interface Props {
  profile: Profile
}

function NoInstallation(props: Props) {
  return (
    <div className="flex flex-col items-center justify-between h-screen">
      <Navbar profile={props.profile} />
      <div className="flex flex-col items-center justify-center space-y-12">
        <h1 className="text-white text-4xl font-bold">
          No organizations configured.
        </h1>
        <a href={installationUrl}>
          <Button>Add an organization</Button>
        </a>
      </div>
      <div />
    </div>
  )
}

export default NoInstallation

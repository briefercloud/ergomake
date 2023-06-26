import React from 'react'

import { Profile } from '../hooks/useProfile'
import AvatarMenu from './AvatarMenu'
import Logo from './Logo'

interface Props {
  profile: Profile
  currentOwner?: string
}
const Navbar = ({ profile, currentOwner }: Props) => {
  return (
    <div className="relative flex top-0 left-0 w-full px-8 py-6 bg-white bg-opacity-5 backdrop-blur-lg drop-shadow-lg justify-between">
      <Logo small />
      <AvatarMenu
        username={profile.username}
        avatar={profile.avatar}
        currentOwner={currentOwner}
      />
    </div>
  )
}

export default Navbar

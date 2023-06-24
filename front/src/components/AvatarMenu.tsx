import { Menu } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/24/solid'
import React from 'react'

import LockIcon from './LockIcon'
import LogoutIcon from './LogoutIcon'

const logoutUrl = `${process.env.REACT_APP_ERGOMAKE_API}/v2/auth/logout?redirectUrl=${window.location.protocol}//${window.location.host}`

interface Props {
  username: string
  avatar: string
  currentOwner?: string
}

function AvatarMenu({ username, avatar, currentOwner }: Props) {
  return (
    <Menu>
      <Menu.Button className="flex items-center space-x-2 px-4 py-2 rounded-md">
        <img
          src={avatar}
          alt="GitHub Logo"
          className="w-8 h-8 rounded-full outline outline-1 outline-primary-300"
        />
        <ChevronDownIcon className="h-3 w-3 text-primary-300" />
      </Menu.Button>
      <Menu.Items className="w-56 absolute right-0 mt-12 mr-8 bg-black rounded-lg border-outcolor border-2">
        <div className="bg-white bg-opacity-10 p-5 rounded-lg text-white font-bold space-y-2">
          <Menu.Item>
            <p className="mb-4">{username}</p>
          </Menu.Item>
          <Menu.Item>
            <a
              href={`/gh/${currentOwner ?? username}/registries`}
              className="flex text-md hover:text-gray-300 items-center"
            >
              <LockIcon className="mr-2" /> Private Registries
            </a>
          </Menu.Item>
          <Menu.Item>
            <a
              href={logoutUrl}
              className="flex text-md hover:text-gray-300 items-center"
            >
              <LogoutIcon className="mr-2" />
              Logout
            </a>
          </Menu.Item>
        </div>
      </Menu.Items>
    </Menu>
  )
}

export default AvatarMenu

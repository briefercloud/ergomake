import { Dialog, Menu, Transition } from '@headlessui/react'
import { ChevronDownIcon } from '@heroicons/react/20/solid'
import {
  Bars3Icon,
  FolderIcon,
  LockClosedIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline'
import { Fragment, useState } from 'react'

import { Profile } from '../../hooks/useProfile'
import Logo from '../components/Logo'
import WebsitePath, { Pages } from '../components/WebsitePath'
import { classNames } from '../utils'

const navigation = [
  { name: 'Repositories', href: '#', icon: FolderIcon, current: true },
  {
    name: 'Private registries',
    href: '#',
    icon: LockClosedIcon,
    current: false,
  },
]
const teams = [
  { id: 1, name: 'Heroicons', href: '#', initial: 'H', current: true },
  { id: 2, name: 'Tailwind Labs', href: '#', initial: 'T', current: false },
  { id: 3, name: 'Workcation', href: '#', initial: 'W', current: false },
  { id: 4, name: 'Add organization', href: '#', initial: '+', current: false },
]

const logoutUrl = `${process.env.REACT_APP_ERGOMAKE_API}/v2/auth/logout?redirectUrl=${window.location.protocol}//${window.location.host}`

const userNavigation = [
  { name: 'Your profile', href: '#' },
  { name: 'Sign out', href: logoutUrl },
]

const sidebarColor = `bg-white shadow`

const DesktopSidebar = () => {
  return (
    <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-72 lg:flex-col">
      <div
        className={classNames(
          sidebarColor,
          `flex grow flex-col gap-y-5 overflow-y-auto px-6 pb-4`
        )}
      >
        <div className="flex h-16 shrink-0 items-center">
          <Logo small />
        </div>
        <nav className="flex flex-1 flex-col">
          <ul className="flex flex-1 flex-col gap-y-12">
            <li>
              <ul className="-mx-2 space-y-1">
                {navigation.map((item) => (
                  <li key={item.name}>
                    <a
                      href={item.href}
                      className={classNames(
                        item.current
                          ? 'text-primary-600 bg-gray-50'
                          : 'text-gray-400 ',
                        'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold hover:text-primary-600 hover:bg-gray-50'
                      )}
                    >
                      <item.icon
                        className={classNames(
                          item.current ? 'text-primary-600' : 'text-gray-400',
                          'h-6 w-6 shrink-0 group-hover:text-primary-600'
                        )}
                        aria-hidden="true"
                      />
                      {item.name}
                    </a>
                  </li>
                ))}
              </ul>
            </li>
            <li>
              <div className="text-xs font-semibold leading-6 text-gray-400">
                Your organizations
              </div>
              <ul className="pt-2">
                {teams.map((team) => (
                  <li key={team.name}>
                    <a
                      href={team.href}
                      className={classNames(
                        team.current ? 'text-primary-600' : 'text-gray-400',
                        'group flex gap-x-3 rounded-md py-1 text-sm leading-6 font-semibold group-hover:text-primary-400'
                      )}
                    >
                      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-primary-400 bg-primary-500 text-[0.625rem] font-medium text-white">
                        {team.initial}
                      </span>
                      <span className="truncate">{team.name}</span>
                    </a>
                  </li>
                ))}
              </ul>
            </li>
          </ul>
        </nav>
      </div>
    </div>
  )
}

type MobileSidebarProps = { sidebarOpen: boolean; closeSidebar: () => void }
const MobileSidebar = ({ sidebarOpen, closeSidebar }: MobileSidebarProps) => {
  return (
    <Transition.Root show={sidebarOpen} as={Fragment}>
      <Dialog
        as="div"
        className="relative z-50 lg:hidden"
        onClose={closeSidebar}
      >
        <Transition.Child
          as={Fragment}
          enter="transition-opacity ease-linear duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="transition-opacity ease-linear duration-300"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-900/80" />
        </Transition.Child>

        <div className="fixed inset-0 flex">
          <Transition.Child
            as={Fragment}
            enter="transition ease-in-out duration-300 transform"
            enterFrom="-translate-x-full"
            enterTo="translate-x-0"
            leave="transition ease-in-out duration-300 transform"
            leaveFrom="translate-x-0"
            leaveTo="-translate-x-full"
          >
            <Dialog.Panel className="relative mr-16 flex w-full max-w-xs flex-1">
              <Transition.Child
                as={Fragment}
                enter="ease-in-out duration-300"
                enterFrom="opacity-0"
                enterTo="opacity-100"
                leave="ease-in-out duration-300"
                leaveFrom="opacity-100"
                leaveTo="opacity-0"
              >
                <div className="absolute left-full top-0 flex w-16 justify-center pt-5">
                  <button
                    type="button"
                    className="-m-2.5 p-2.5"
                    onClick={closeSidebar}
                  >
                    <span className="sr-only">Close sidebar</span>
                    <XMarkIcon
                      className="h-6 w-6 text-white"
                      aria-hidden="true"
                    />
                  </button>
                </div>
              </Transition.Child>

              <div
                className={classNames(
                  sidebarColor,
                  'flex grow flex-col gap-y-5 overflow-y-auto px-6 pb-4'
                )}
              >
                <div className="flex h-16 shrink-0 items-center">
                  <Logo small />
                </div>
                <nav className="flex flex-1 flex-col">
                  <ul className="flex flex-1 flex-col gap-y-7">
                    <li>
                      <ul className="-mx-2 space-y-1">
                        {navigation.map((item) => (
                          <li key={item.name}>
                            <a
                              href={item.href}
                              className={classNames(
                                item.current
                                  ? 'bg-primary-700 text-white'
                                  : 'text-primary-200 hover:text-white hover:bg-primary-700',
                                'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold'
                              )}
                            >
                              <item.icon
                                className={classNames(
                                  item.current
                                    ? 'text-white'
                                    : 'text-primary-200 group-hover:text-white',
                                  'h-6 w-6 shrink-0'
                                )}
                                aria-hidden="true"
                              />
                              {item.name}
                            </a>
                          </li>
                        ))}
                      </ul>
                    </li>
                    <li>
                      <div className="text-xs font-semibold leading-6 text-primary-200">
                        Your teams
                      </div>
                      <ul className="-mx-2 mt-2 space-y-1">
                        {teams.map((team) => (
                          <li key={team.name}>
                            <a
                              href={team.href}
                              className={classNames(
                                team.current
                                  ? 'text-primary-400'
                                  : 'text-gray-400 hover:text-white',
                                'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold'
                              )}
                            >
                              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-primary-400 bg-primary-500 text-[0.625rem] font-medium text-white">
                                {team.initial}
                              </span>
                              <span className="truncate">{team.name}</span>
                            </a>
                          </li>
                        ))}
                      </ul>
                    </li>
                  </ul>
                </nav>
              </div>
            </Dialog.Panel>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  )
}

type TopNavbarProps = {
  profile: Profile
  openSidebar: () => void
  pages: Pages[]
}

const TopNavbar = ({ profile, openSidebar, pages }: TopNavbarProps) => {
  return (
    <div className="sticky top-0 z-40 flex h-16 shrink-0 items-center gap-x-4 border-b border-gray-200 bg-white px-4 shadow-sm md:shadow-none md:border-0 md:bg-transparent sm:gap-x-6 sm:px-6 lg:px-8 flex justify-between">
      <button
        type="button"
        className="-m-2.5 p-2.5 text-gray-700 lg:hidden"
        onClick={openSidebar}
      >
        <span className="sr-only">Open sidebar</span>
        <Bars3Icon className="h-6 w-6" aria-hidden="true" />
      </button>

      <div className="invisible md:visible">
        <WebsitePath pages={pages} />

        <div className="h-6 w-px bg-gray-900/10 lg:hidden" aria-hidden="true" />
      </div>

      <div className="flex gap-x-4 self-stretch lg:gap-x-6">
        <div className="flex items-center gap-x-4 lg:gap-x-6">
          <Menu as="div" className="relative">
            <Menu.Button className="-m-1.5 flex items-center p-1.5">
              <span className="sr-only">Open user menu</span>
              <img
                className="h-8 w-8 rounded-full bg-gray-50"
                src={profile.avatar}
                alt=""
              />
              <span className="hidden lg:flex lg:items-center">
                <ChevronDownIcon
                  className="ml-2 h-5 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </span>
            </Menu.Button>
            <Transition
              as={Fragment}
              enter="transition ease-out duration-100"
              enterFrom="transform opacity-0 scale-95"
              enterTo="transform opacity-100 scale-100"
              leave="transition ease-in duration-75"
              leaveFrom="transform opacity-100 scale-100"
              leaveTo="transform opacity-0 scale-95"
            >
              <Menu.Items className="absolute right-0 z-10 mt-2.5 w-32 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none">
                {userNavigation.map((item) => (
                  <Menu.Item key={item.name}>
                    {({ active }) => (
                      <a
                        href={item.href}
                        className={classNames(
                          active ? 'bg-gray-50' : '',
                          'block px-3 py-1 text-sm leading-6 text-gray-900'
                        )}
                      >
                        {item.name}
                      </a>
                    )}
                  </Menu.Item>
                ))}
              </Menu.Items>
            </Transition>
          </Menu>
        </div>
      </div>
    </div>
  )
}

type LayoutProps = {
  children: React.ReactNode
  profile: Profile
  pages: Pages[]
}

const Layout = ({ profile, children, pages }: LayoutProps) => {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <>
      <div>
        <DesktopSidebar />

        <MobileSidebar
          sidebarOpen={sidebarOpen}
          closeSidebar={() => setSidebarOpen(false)}
        />

        <div className="lg:pl-72">
          <TopNavbar
            pages={pages}
            profile={profile}
            openSidebar={() => setSidebarOpen(true)}
          />

          <div className="invisible md:visible flex items-center">
            <div className="w-full border-t border-gray-200" />
          </div>

          <main>{children}</main>
        </div>
      </div>
    </>
  )
}

export default Layout

import {
  ChevronRightIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/solid'
import React from 'react'
import { Link } from 'react-router-dom'

import Alert from '../components/Alert'
import Background from '../components/Background'
import Button from '../components/Button'
import CheckCircleIcon from '../components/CheckCircleIcon'
import Loading from '../components/Loading'
import Navbar from '../components/Navbar'
import Pane from '../components/Pane'
import Select from '../components/Select'
import TextInput from '../components/TextInput'
import XCircleIcon from '../components/XCircleIcon'
import { Owner } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Repo } from '../hooks/useRepo'

interface Props {
  profile: Profile
  owners: Owner[]
  owner: Owner
  onChangeOwner: (org: Owner) => void
  repositories: Repo[]
  loadingRepos: boolean
  search: string
  onChangeSearch: (search: string) => void
  hasProjects: boolean
}

const installationUrl = 'https://github.com/apps/ergomake/installations/new'

function Projects(props: Props) {
  return (
    <Background>
      <Navbar profile={props.profile} currentOwner={props.owner.login} />
      <div className="px-20 pt-8 space-y-8">
        {!props.owner.isPaying && <Alert owner={props.owner.login} />}
        <Pane className="w-full space-y-8">
          <div className="flex justify-between">
            <h1 className="text-2xl font-bold">Your Projects</h1>
            <a href={installationUrl}>
              <Button>Add project</Button>
            </a>
          </div>
          <div className="flex justify-between">
            <Select
              values={props.owners}
              value={props.owner}
              onChange={props.onChangeOwner}
              cta={<a href={installationUrl}>Add an organization</a>}
              small
            >
              {(org) => (
                <div className="flex items-center p-2">
                  <img
                    src={org.avatar}
                    alt={`Avatar of ${org.login}`}
                    className="w-6 h-6 mr-2 rounded-full"
                  />
                  {org.login}
                </div>
              )}
            </Select>
            <TextInput
              value={props.search}
              onChange={props.onChangeSearch}
              placeholder="Search"
              Icon={MagnifyingGlassIcon}
              className="w-full ml-4"
            />
          </div>
        </Pane>
        {props.loadingRepos && (
          <div className="flex justify-center">
            <Loading />
          </div>
        )}
        <ul className="w-full">
          {props.repositories.map((repo) => (
            <li key={repo.name} className="mb-3">
              <Link to={`/gh/${props.owner.login}/repos/${repo.name}/`}>
                <Pane className="w-full flex justify-between items-center">
                  <p className="text-xl font-bold">{repo.name}</p>
                  <div className="flex items-center">
                    {repo.isInstalled ? (
                      <>
                        <CheckCircleIcon />
                        <p className="ml-2 text-sm font-bold text-primary-300">
                          {repo.environmentCount} environment
                          {repo.environmentCount !== 1 ? 's' : ''}
                        </p>
                      </>
                    ) : (
                      <>
                        <XCircleIcon />
                        <p className="ml-2 text-sm font-bold text-red-400">
                          Not configured
                        </p>
                      </>
                    )}
                    <ChevronRightIcon className="ml-8 w-4 h-4 text-primary-500" />
                  </div>
                </Pane>
              </Link>
            </li>
          ))}
        </ul>
        <div className="flex flex-col items-center justify-center w-full h-full py-8">
          <span className="text-xl text-gray-500">
            {props.hasProjects ? "Can't see your project?" : 'No projects yet.'}{' '}
            <a
              className="text-primary-500 hover:underline hover:text-primary-400"
              href={installationUrl}
            >
              Add a project.
            </a>
          </span>
        </div>
      </div>
    </Background>
  )
}

export default Projects

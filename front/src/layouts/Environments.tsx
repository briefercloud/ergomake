import {
  ChevronLeftIcon,
  ChevronRightIcon,
  MagnifyingGlassIcon,
  MoonIcon,
} from '@heroicons/react/24/solid'
import React from 'react'
import { Link } from 'react-router-dom'

import AlertCircleIcon from '../components/AlertCircleIcon'
import Background from '../components/Background'
import Button from '../components/Button'
import CheckCircleIcon from '../components/CheckCircleIcon'
import Loading from '../components/Loading'
import ManageVariablesModal from '../components/ManageVariablesModal'
import Navbar from '../components/Navbar'
import Pane from '../components/Pane'
import Select from '../components/Select'
import TextInput from '../components/TextInput'
import XCircleIcon from '../components/XCircleIcon'
import { Environment } from '../hooks/useEnvironment'
import { Owner } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Repo } from '../hooks/useRepo'
import { Variable } from '../hooks/useVariables'

interface Props {
  profile: Profile
  owner: Owner

  repos: Repo[]
  repo: Repo
  onChangeRepo: (repo: Repo) => void

  loadingEnvironments: boolean
  environments: Environment[]

  search: string
  onChangeSearch: (search: string) => void

  variables: Variable[]
  onManageVariables: () => void
  managingVariables: boolean
  onDiscardVariableManagement: () => void
  onUpdateVariables: (vars: Variable[]) => void
  updatingVariables: boolean
}
function Environments(props: Props) {
  return (
    <Background>
      <Navbar profile={props.profile} currentOwner={props.owner.login} />
      <div className="px-20 pt-8 space-y-8">
        <Link
          className="flex items-center text-primary-500 font-bold text-sm"
          to={`/gh/${props.owner.login}`}
        >
          <ChevronLeftIcon className="w-5 h-5 text-primary-500 mr-1" />
          Back to projects
        </Link>
        <Pane>
          <div className="flex justify-between">
            <h2 className="flex items-center text-white font-bold text-sm">
              <img
                className="mr-2 w-8 h-8 rounded-full"
                src={props.owner.avatar}
                alt={props.owner.login}
              />
              {props.owner.login}
              <span className="mx-2 text-primary-500">/</span>
              {props.repo.name}
            </h2>
            <Button
              onClick={props.onManageVariables}
              disabled={props.updatingVariables}
            >
              Manage variables
            </Button>
          </div>
          <h1 className="mt-4 mb-8 text-white font-bold text-4xl">
            {props.repo.name}
          </h1>

          <div className="flex justify-between">
            <div className="flex-2">
              <Select
                values={props.repos}
                value={props.repo}
                onChange={props.onChangeRepo}
                small
              >
                {(repo) => (
                  <div className="flex items-center p-2">{repo.name}</div>
                )}
              </Select>
            </div>
            <div className="flex-1 ml-4">
              <TextInput
                value={props.search}
                onChange={props.onChangeSearch}
                placeholder="Search"
                Icon={MagnifyingGlassIcon}
              />
            </div>
          </div>
        </Pane>
        {props.loadingEnvironments && (
          <div className="flex justify-center">
            <Loading />
          </div>
        )}
        <ul className="w-full mt-3">
          {props.environments.map((env) => (
            <li key={env.id} className="mb-3">
              <Link
                to={`/gh/${props.owner.login}/repos/${props.repo.name}/envs/${env.id}/`}
              >
                <Pane className="w-full flex justify-between items-center">
                  <div className="flex-1">
                    <p className="text-sm">Environment</p>
                    <p className="text-xl font-bold">{env.branch}</p>
                  </div>
                  <div>
                    <p className="text-sm">Source</p>
                    <p className="text-xl font-bold">
                      {env.source.toUpperCase()}
                    </p>
                  </div>
                  <div className="flex flex-1 items-center justify-end">
                    {env.status === 'success' ? (
                      <>
                        <CheckCircleIcon />
                        <p className="ml-2 text-sm font-bold text-primary-300">
                          Success
                        </p>
                      </>
                    ) : env.status === 'degraded' ? (
                      <>
                        <XCircleIcon />
                        <p className="ml-2 text-sm font-bold text-red-400">
                          Failed
                        </p>
                      </>
                    ) : env.status === 'limited' ? (
                      <>
                        <XCircleIcon />
                        <p className="ml-2 text-sm font-bold text-red-400">
                          Limited
                        </p>
                      </>
                    ) : env.status === 'stale' ? (
                      <>
                        <MoonIcon className="w-4 h-4" />
                        <p className="ml-2 text-sm font-bold text-gray-300">
                          Sleeping
                        </p>
                      </>
                    ) : env.status === 'pending' ? (
                      <>
                        <AlertCircleIcon />
                        <p className="ml-2 text-sm font-bold text-gray-300">
                          Building
                        </p>
                      </>
                    ) : null}
                    <ChevronRightIcon className="ml-8 w-4 h-4 text-primary-500" />
                  </div>
                </Pane>
              </Link>
            </li>
          ))}
        </ul>
        <ManageVariablesModal
          open={props.managingVariables}
          variables={props.variables}
          onDiscard={props.onDiscardVariableManagement}
          onUpdate={props.onUpdateVariables}
          updating={props.updatingVariables}
        />
      </div>
    </Background>
  )
}

export default Environments

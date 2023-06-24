import { ChevronLeftIcon } from '@heroicons/react/24/solid'
import React from 'react'
import { Link } from 'react-router-dom'

import AddRegistryModal from '../components/AddRegistryModal'
import Background from '../components/Background'
import Button from '../components/Button'
import Loading from '../components/Loading'
import Navbar from '../components/Navbar'
import Pane from '../components/Pane'
import Select from '../components/Select'
import TrashIcon from '../components/TrashIcon'
import { Owner } from '../hooks/useOwners'
import { Profile } from '../hooks/useProfile'
import { Registry } from '../hooks/useRegistries'

export const showProvider = (provider: string): string => {
  switch (provider) {
    case 'ecr':
      return 'Amazon Elastic Container Registry'
    case 'gcr':
      return 'Google Cloud Registry'
    case 'hub':
      return 'Dockerhub'
    default:
      return provider
  }
}

interface Props {
  profile: Profile
  owner: Owner
  owners: Owner[]
  onChangeOwner: (owner: Owner) => void
  registries: Registry[]
  loading: boolean
  adding: boolean
  onOpenAdding: () => void
  onCloseAdding: () => void
  onDelete: (reg: Registry) => void
}
function Registries(props: Props) {
  return (
    <Background>
      <Navbar profile={props.profile} currentOwner={props.owner.login} />
      <div className="px-20 pt-8 space-y-8">
        <Link
          className="flex items-center text-primary-500 font-bold text-sm"
          to={`/gh/${props.owner.login}`}
        >
          <ChevronLeftIcon className="w-5 h-5 text-primary-500 mr-1" />
          Back
        </Link>
        <Pane className="w-full space-y-8">
          <div className="flex justify-between">
            <h1 className="text-2xl font-bold">Private Registries</h1>

            <Button onClick={props.onOpenAdding}>Add registry</Button>
          </div>
          <Select
            values={props.owners}
            value={props.owner}
            onChange={props.onChangeOwner}
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
        </Pane>
        {props.loading && (
          <div className="flex justify-center">
            <Loading />
          </div>
        )}
        <ul className="w-full mt-3">
          {props.registries.map((reg) => (
            <li key={reg.url} className="mb-3">
              <Pane className="w-full flex justify-between items-center space-x-4">
                <div>
                  <p className="text-sm">URL</p>
                  <p className="text-xl font-bold">{reg.url}</p>
                </div>
                <div>
                  <p className="text-sm">Provider</p>
                  <p className="text-xl font-bold">
                    {showProvider(reg.provider)}
                  </p>
                </div>
                <Button variant="text" onClick={() => props.onDelete(reg)}>
                  <TrashIcon />
                </Button>
              </Pane>
            </li>
          ))}
        </ul>
        <AddRegistryModal
          open={props.adding}
          onClose={props.onCloseAdding}
          owner={props.owner.login}
        />
      </div>
    </Background>
  )
}

export default Registries

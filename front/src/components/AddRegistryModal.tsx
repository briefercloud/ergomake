import { equals } from 'ramda'
import React, { useCallback, useEffect, useState } from 'react'

import { useAddRegistry } from '../hooks/useAddRegistry'
import { isLoading, isSuccess } from '../hooks/useHTTPRequest'
import { RegistryProvider } from '../hooks/useRegistries'
import { showProvider } from '../layouts/Registries'
import Button from './Button'
import Modal from './Modal'
import Pane from './Pane'
import Select from './Select'
import TextInput from './TextInput'

const providers = ['ecr']
type Region =
  | 'us-east-1'
  | 'us-east-2'
  | 'us-west-1'
  | 'us-west-2'
  | 'af-south-1'
  | 'ap-east-1'
  | 'ap-south-2'
  | 'ap-southeast-3'
  | 'ap-southeast-4'
  | 'ap-south-1'
  | 'ap-northeast-3'
  | 'ap-northeast-2'
  | 'ap-southeast-1'
  | 'ap-southeast-2'
  | 'ap-northeast-1'
  | 'ca-central-1'
  | 'eu-central-1'
  | 'eu-west-1'
  | 'eu-west-2'
  | 'eu-south-1'
  | 'eu-west-3'
  | 'eu-south-2'
  | 'eu-north-1'
  | 'eu-central-2'
  | 'me-south-1'
  | 'me-central-1'
  | 'sa-east-1'
const regions: Region[] = [
  'us-east-1',
  'us-east-2',
  'us-west-1',
  'us-west-2',
  'af-south-1',
  'ap-east-1',
  'ap-south-2',
  'ap-southeast-3',
  'ap-southeast-4',
  'ap-south-1',
  'ap-northeast-3',
  'ap-northeast-2',
  'ap-southeast-1',
  'ap-southeast-2',
  'ap-northeast-1',
  'ca-central-1',
  'eu-central-1',
  'eu-west-1',
  'eu-west-2',
  'eu-south-1',
  'eu-west-3',
  'eu-south-2',
  'eu-north-1',
  'eu-central-2',
  'me-south-1',
  'me-central-1',
  'sa-east-1',
]

type Validation = {
  url: boolean
  accessKeyId: boolean
  secretAccessKey: boolean
}
const isValid = (v: Validation): boolean =>
  v.url && v.accessKeyId && v.secretAccessKey

interface Props {
  open: boolean
  onClose: () => void
  owner: string
}
function AddRegistryModal({ open, onClose, owner }: Props) {
  const [url, setUrl] = useState('')
  const [accessKeyId, setAccessKeyId] = useState('')
  const [secretAccessKey, setSecretAccessKey] = useState('')
  const [region, setRegion] = useState<Region>('us-east-1')
  const [provider, setProvider] = useState('ecr')
  const [validation, setValidation] = useState<Validation>({
    url: true,
    accessKeyId: true,
    secretAccessKey: true,
  })

  const reset = useCallback(() => {
    setUrl('')
    setAccessKeyId('')
    setSecretAccessKey('')
    setRegion('us-east-1')
    setProvider('ecr')
    setValidation({
      url: true,
      accessKeyId: true,
      secretAccessKey: true,
    })
  }, [
    setUrl,
    setAccessKeyId,
    setSecretAccessKey,
    setRegion,
    setProvider,
    setValidation,
  ])
  const onDismiss = useCallback(() => {
    onClose()
    setTimeout(reset, 100)
  }, [reset, onClose])

  const [res, add] = useAddRegistry(owner)
  const updating =
    res._tag !== 'pristine' &&
    (isLoading(res) || (isSuccess(res) && res.loading))
  const onSave = useCallback(() => {
    if (res._tag !== 'pristine' && isLoading(res)) {
      return
    }

    const validation: Validation = {
      url: url.trim() !== '',
      accessKeyId: accessKeyId.trim() !== '',
      secretAccessKey: secretAccessKey.trim() !== '',
    }

    setValidation(validation)
    if (!isValid(validation)) {
      return
    }

    add({
      url,
      provider: provider as RegistryProvider,
      credentials: JSON.stringify({ accessKeyId, secretAccessKey, region }),
    })
  }, [res, url, provider, accessKeyId, secretAccessKey, region, add])

  useEffect(() => {
    if (res._tag === 'pristine') {
      return
    }

    if (isSuccess(res)) {
      onDismiss()
    }
  }, [res, onDismiss])

  useEffect(() => {
    if (isValid(validation)) {
      return
    }

    const nextVal = {
      url: url.trim() !== '',
      accessKeyId: accessKeyId.trim() !== '',
      secretAccessKey: secretAccessKey.trim() !== '',
    }

    if (equals(validation, nextVal)) {
      return
    }

    setValidation(nextVal)
  }, [validation, url, accessKeyId, secretAccessKey])

  return (
    <Modal open={open} onDismiss={onDismiss}>
      <Pane>
        <h1 className="font-bold text-lg mb-4">Add registry</h1>
        <div className="space-y-4">
          <TextInput
            label="URL"
            value={url}
            onChange={setUrl}
            error={!validation.url}
          />
          <Select
            label="Provider"
            values={providers}
            value={provider}
            onChange={setProvider}
            fullWidth
          >
            {(provider) => <span>{showProvider(provider)}</span>}
          </Select>
          <TextInput
            label="Access Key Id"
            value={accessKeyId}
            onChange={setAccessKeyId}
            error={!validation.accessKeyId}
          />
          <TextInput
            label="Secret Access Key"
            value={secretAccessKey}
            onChange={setSecretAccessKey}
            error={!validation.secretAccessKey}
          />
          <Select
            label="Region"
            values={regions}
            value={region}
            onChange={setRegion}
            fullWidth
          >
            {(region) => <span>{region}</span>}
          </Select>
          <div className="flex mt-2 space-x-4 justify-end">
            <Button onClick={onSave} disabled={updating} loading={updating}>
              Save
            </Button>
            <Button variant="text" onClick={onDismiss} disabled={updating}>
              Close
            </Button>
          </div>
        </div>
      </Pane>
    </Modal>
  )
}

export default AddRegistryModal

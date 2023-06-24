import { TrashIcon } from '@heroicons/react/24/solid'
import React, { useCallback, useEffect, useState } from 'react'

import { Variable } from '../hooks/useVariables'
import Button from './Button'
import Modal from './Modal'
import Pane from './Pane'
import TextInput from './TextInput'

interface Props {
  open: boolean
  variables: Variable[]
  onDiscard: () => void
  onUpdate: (variables: Variable[]) => void
  updating: boolean
}
function ManageVariablesModal({
  open,
  variables,
  onDiscard,
  onUpdate,
  updating,
}: Props) {
  const [state, setState] = useState(variables)
  useEffect(() => {
    setState(variables)
  }, [variables])

  const onAdd = useCallback(() => {
    setState((vars) => [...vars, { name: '', value: '' }])
  }, [setState])

  const onRemove = useCallback(
    (i: number) => {
      setState((vars) => [...vars.slice(0, i), ...vars.slice(i + 1)])
    },
    [setState]
  )

  const onChangeName = useCallback(
    (name: string, i: number) => {
      setState((vars) =>
        vars.map((v, j) => ({ ...v, name: i === j ? name : v.name }))
      )
    },
    [setState]
  )

  const onChangeValue = useCallback(
    (value: string, i: number) => {
      setState((vars) =>
        vars.map((v, j) => ({ ...v, value: i === j ? value : v.value }))
      )
    },
    [setState]
  )

  const handleOnDiscard = useCallback(() => {
    if (updating) {
      return
    }

    setState(variables)
    onDiscard()
  }, [updating, variables, onDiscard, setState])

  const onSave = useCallback(() => {
    onUpdate(state.filter((v) => v.name.trim() !== ''))
  }, [onUpdate, state])

  return (
    <Modal open={open} onDismiss={handleOnDiscard}>
      <Pane>
        <div className="flex justify-between items-center mb-4">
          <h1 className="font-bold text-lg mb-4">Variables</h1>
        </div>
        <ul className="space-y-2">
          {state.map(({ name, value }, i) => (
            <li key={i}>
              <div className="flex justify-between items-center">
                <TextInput
                  label="Name"
                  value={name}
                  onChange={(name) => onChangeName(name, i)}
                  disabled={updating}
                />
                <TextInput
                  label="Value"
                  value={value}
                  onChange={(value) => onChangeValue(value, i)}
                  disabled={updating}
                />
                <Button
                  variant="text"
                  disabled={updating}
                  onClick={() => onRemove(i)}
                >
                  <TrashIcon className="w-5 h-5" />
                </Button>
              </div>
            </li>
          ))}
        </ul>
        <Button className="mt-2" onClick={onAdd}>
          Add
        </Button>
        <div className="flex space-x-4 justify-end">
          <Button onClick={onSave} disabled={updating}>
            Save
          </Button>

          <Button variant="text" onClick={onDiscard}>
            Close
          </Button>
        </div>
      </Pane>
    </Modal>
  )
}

export default ManageVariablesModal

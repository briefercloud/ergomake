import React, { useCallback, useEffect, useMemo, useState } from 'react'

import TableInput from '../components/TableInput'
import { isLoading, isSuccess } from '../hooks/useHTTPRequest'
import { Variable, useVariables } from '../hooks/useVariables'

const labels = ['Name', 'Value']
const placeholders = ['EXAMPLE_VAR', 'value123']

const envVarRegex = /^[a-zA-Z_][a-zA-Z0-9_]*$/

interface Props {
  owner: string
  repo: string
}
function VariablesInput(props: Props) {
  const [res, onUpdate] = useVariables(props.owner, props.repo)

  const [variables, setVariables] = useState<Variable[]>([])
  useEffect(() => {
    if (isSuccess(res) && !res.refreshing) {
      setVariables(res.body)
    }
  }, [res])

  const values = useMemo(
    () => variables.map((v) => [v.name, v.value]),
    [variables]
  )

  const onAdd = useCallback(
    ([name, value]: string[]) => {
      if (!name || !value) {
        return false
      }

      if (!envVarRegex.test(name)) {
        return false
      }

      setVariables((vars) => [{ name, value }].concat(vars))
      return true
    },
    [setVariables]
  )
  const onRemove = useCallback(
    (i: number) => {
      setVariables((vars) => [...vars.slice(0, i), ...vars.slice(i + 1)])
    },
    [setVariables]
  )
  const onSave = useCallback(() => {
    onUpdate(variables)
  }, [variables, onUpdate])

  return (
    <TableInput
      values={values}
      loading={isLoading(res) || (isSuccess(res) && res.refreshing)}
      cells={2}
      labels={labels}
      placeholders={placeholders}
      onAdd={onAdd}
      onRemove={onRemove}
      onSave={onSave}
      saveLabel="Save to apply environment variables."
    />
  )
}

export default VariablesInput

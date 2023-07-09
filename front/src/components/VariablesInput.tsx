import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { isLoading, isSuccess } from '../hooks/useHTTPRequest'
import { Variable, useVariables } from '../hooks/useVariables'
import TableInput from '../v2/components/TableInput'

const labels = ['Name', 'Value']
const placeholders = ['EXAMPLE_VAR', 'value123']

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
      console.log({ name, value })
      if (!name || !value) {
        return
      }

      setVariables((vars) => [{ name, value }].concat(vars))
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
    />
  )
}

export default VariablesInput

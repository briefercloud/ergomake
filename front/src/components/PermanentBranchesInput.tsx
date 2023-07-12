import { useCallback, useEffect, useMemo, useState } from 'react'

import TableInput from '../components/TableInput'
import { isLoading, isSuccess } from '../hooks/useHTTPRequest'
import { usePermanentBranches } from '../hooks/usePermanentBranches'

const labels = ['Name']
const placeholders = ['main']

const branchNameRegex = /^[a-zA-Z_][a-zA-Z0-9_]*$/

interface Props {
  owner: string
  repo: string
}
function PermanentBranchesInput({ owner, repo }: Props) {
  const [res, onUpdate] = usePermanentBranches(owner, repo)

  const [branches, setBranches] = useState<string[]>([])
  useEffect(() => {
    if (isSuccess(res) && !res.refreshing) {
      setBranches(res.body.map((r) => r.name))
    }
  }, [res])

  const values = useMemo(() => branches.map((b) => [b]), [branches])

  const onAdd = useCallback(
    ([branch]: string[]) => {
      if (!branch) {
        return false
      }

      if (!branchNameRegex.test(branch)) {
        return false
      }

      setBranches((branches) => [branch].concat(branches))
      return true
    },
    [setBranches]
  )
  const onRemove = useCallback(
    (i: number) => {
      setBranches((branches) => [
        ...branches.slice(0, i),
        ...branches.slice(i + 1),
      ])
    },
    [setBranches]
  )
  const onSave = useCallback(() => {
    onUpdate({ branches })
  }, [branches, onUpdate])

  return (
    <TableInput
      values={values}
      loading={isLoading(res) || (isSuccess(res) && res.refreshing)}
      cells={1}
      labels={labels}
      placeholders={placeholders}
      onAdd={onAdd}
      onRemove={onRemove}
      onSave={onSave}
      saveLabel="Save to update your permanent branches."
    />
  )
}

export default PermanentBranchesInput

import { Combobox } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import { TrashIcon } from '@heroicons/react/24/outline'
import classNames from 'classnames'
import * as R from 'ramda'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { isLoading, isSuccess, map, orElse } from '../hooks/useHTTPRequest'
import { useRepo } from '../hooks/useRepo'
import { Variable, useVariables } from '../hooks/useVariables'
import Button from './Button'
import Input from './Input'

const envVarRegex = /^[a-zA-Z_][a-zA-Z0-9_]*$/

interface Props {
  owner: string
  repo: string
}
function VariablesInput(props: Props) {
  const repo = useRepo(props.owner, props.repo)
  const [res, onUpdate] = useVariables(props.owner, props.repo)

  const [variables, setVariables] = useState<Variable[]>([])
  useEffect(() => {
    if (isSuccess(res) && !res.refreshing) {
      setVariables(res.body)
    }
  }, [res])
  const [current, setCurrent] = useState<Variable>({
    name: '',
    value: '',
    branch: null,
  })
  const branches = orElse(
    map(repo, (r) => r?.branches ?? []),
    []
  ).filter((b) => {
    if (!current.branch || current.branch === '') {
      return true
    }

    return b.toLowerCase().includes(current.branch?.toLowerCase())
  })

  const onChangeCurrentName = useCallback(
    (name: string) => {
      setCurrent((c) => ({ ...c, name }))
    },
    [setCurrent]
  )
  const onChangeCurrentValue = useCallback(
    (value: string) => {
      setCurrent((c) => ({ ...c, value }))
    },
    [setCurrent]
  )
  const onChangeCurrentBranch = useCallback(
    (branch: string) => {
      setCurrent((c) => ({ ...c, branch: branch === '' ? null : branch }))
    },
    [setCurrent]
  )

  const onAdd: React.FormEventHandler = useCallback(
    (e) => {
      e.preventDefault()

      if (!envVarRegex.test(current.name)) {
        return
      }

      const newVar: Variable = {
        name: current.name.trim(),
        value: current.value.trim(),
        branch: current.branch?.trim() ?? null,
      }
      if (newVar.branch === '') {
        newVar.branch = null
      }

      if (!newVar.name || !newVar.value) {
        return
      }

      setVariables((vars) => [newVar].concat(vars))
      setCurrent({ name: '', value: '', branch: null })
      return
    },
    [current, setCurrent, setVariables]
  )

  const onRemove = useCallback(
    (name: string, value: string, branches: (string | null)[]) => () => {
      const newVars = variables.filter((v) => {
        const sameNameVal = v.name === name && v.value === value
        if (!sameNameVal) {
          return true
        }

        const sameBranch = branches.some((b) => b === v.branch)

        return !sameBranch
      })

      setVariables(newVars)
    },
    [variables, setVariables]
  )

  const onSave = useCallback(() => {
    onUpdate(variables)
  }, [variables, onUpdate])

  const loading = isLoading(res) || (isSuccess(res) && res.refreshing)

  const body = useMemo(() => {
    const byName = R.groupBy((v) => v.name, variables)
    const rows = R.flatten(
      Object.entries(byName).map(([name, vars]) => {
        const byValue = R.groupBy((v) => v.value, vars)
        return Object.entries(byValue).map(([value, vars]) => {
          const branches = vars.map((v) => v.branch)

          return { name, value, branches }
        })
      })
    )

    return R.sortWith(
      [R.ascend((r) => r.name), R.ascend((r) => r.value)],
      rows
    ).map(({ name, value, branches }, i) => (
      <div className="flex items-center w-full h-16 border-b dark:border-neutral-800 border-gray-200 px-4 sm:px-6 lg:px-8">
        <span
          className={classNames(
            'flex-1 grow whitespace-nowrap py-4 text-sm font-medium text-gray-500 dark:text-neutral-500'
          )}
        >
          {name}
        </span>
        <span
          className={classNames(
            'flex-1 grow whitespace-nowrap py-4 text-sm font-medium text-gray-500 dark:text-neutral-500'
          )}
        >
          {value}
        </span>
        <span
          className={classNames(
            'flex-1 grow whitespace-nowrap py-4 text-sm font-medium space-x-1'
          )}
        >
          {R.sortWith(
            [R.ascend((b) => (b === null ? 0 : 1)), R.ascend((b) => b ?? '')],
            branches
          ).map((branch) => (
            <span
              key={branch}
              className={classNames(
                'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                {
                  'bg-green-50  text-green-700 ring-green-600/20 dark:bg-green-500/10 dark:text-green-400 dark:ring-green-500/20':
                    branch === null,
                  'bg-blue-50 text-blue-700 ring-blue-700/10 dark:bg-blue-400/10 dark:text-blue-400 dark:ring-blue-400/30':
                    branch !== null,
                }
              )}
            >
              {branch ?? 'All'}
            </span>
          ))}
        </span>
        <span className="whitespace-nowrap py-4 flex items-center justify-center w-16">
          <button onClick={onRemove(name, value, branches)} disabled={loading}>
            <TrashIcon
              className={classNames('h-5 w-5 flex-shrink-0', {
                'text-red-600 dark:text-red-400 hover:text-red-400 dark:hover:text-red-300 hover:cursor-pointer':
                  !loading,
                'text-gray-600': loading,
              })}
              aria-hidden="true"
            />
          </button>
        </span>
      </div>
    ))
  }, [loading, onRemove, variables])

  return (
    <div className="w-full flex flex-col h-full overflow-y-hidden">
      <form onSubmit={onAdd}>
        <div className="w-full flex pt-4 px-4 sm:px-6 lg:px-8">
          <span className="flex-1 text-sm grow flex-1 text-left font-semibold text-gray-900 dark:text-gray-300">
            Name
          </span>
          <span className="flex-1 text-sm grow flex-1 text-left font-semibold text-gray-900 dark:text-gray-300">
            Value
          </span>
          <span className="flex-1 text-sm grow flex-1 text-left font-semibold text-gray-900 dark:text-gray-300">
            Branch (optional)
          </span>
          <span className="w-16"></span>
        </div>

        <div className="w-full flex border-b border-gray-200 dark:border-neutral-800 py-4 px-4 sm:px-6 lg:px-8">
          <div className="flex-1 grow">
            <Input
              label="Name"
              onChange={onChangeCurrentName}
              value={current.name}
              placeholder="EXAMPLE_VAR"
              disabled={loading}
            />
          </div>
          <div className="flex-1 grow">
            <Input
              label="Value"
              onChange={onChangeCurrentValue}
              value={current.value}
              placeholder="value123"
              disabled={loading}
            />
          </div>
          <div className="flex-1 grow">
            <Combobox
              as="div"
              value={current.branch}
              onChange={onChangeCurrentBranch}
              className="pr-4"
            >
              <div className="relative">
                <Combobox.Input
                  className={classNames(
                    'block w-full rounded-md border-0 py-1.5 text-gray-900 dark:text-gray-200 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-neutral-800 placeholder:text-gray-400 dark:placeholder:text-gray-600 focus:ring-2 focus:ring-inset dark:focus:ring-primary-800 focus:ring-primary-600 sm:text-sm sm:leading-6 dark:bg-neutral-950/30',
                    { 'bg-gray-100 dark:bg-neutral-700': loading }
                  )}
                  onChange={(event) =>
                    onChangeCurrentBranch(event.target.value)
                  }
                  placeholder="New or existing branch"
                />
                <Combobox.Button className="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none">
                  <ChevronUpDownIcon
                    className="h-5 w-5 text-gray-400"
                    aria-hidden="true"
                  />
                </Combobox.Button>

                {branches.length > 0 && (
                  <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm dark:bg-neutral-800">
                    {branches.map((branch) => (
                      <Combobox.Option
                        key={branch}
                        value={branch}
                        className={({ active }) =>
                          classNames(
                            'relative cursor-default select-none py-2 pl-3 pr-9  dark:bg-neutral-800 dark:text-white',
                            active
                              ? 'dark:bg-primary-600 bg-primary-600 text-white'
                              : 'text-gray-900'
                          )
                        }
                      >
                        {({ active, selected }) => (
                          <>
                            <span
                              className={classNames(
                                'block truncate',
                                selected && 'font-semibold'
                              )}
                            >
                              {branch}
                            </span>

                            {selected && (
                              <span
                                className={classNames(
                                  'absolute inset-y-0 right-0 flex items-center pr-4',
                                  active ? 'text-white' : 'text-indigo-600'
                                )}
                              >
                                <CheckIcon
                                  className="h-5 w-5"
                                  aria-hidden="true"
                                />
                              </span>
                            )}
                          </>
                        )}
                      </Combobox.Option>
                    ))}
                  </Combobox.Options>
                )}
              </div>
            </Combobox>
          </div>
          <div className="w-16">
            <Button className="w-full" type="submit" disabled={loading}>
              Add
            </Button>
          </div>
        </div>
      </form>

      <div className="h-full overflow-y-auto">{body}</div>
      <div className="flex bg-gray-200 dark:bg-neutral-800 dark:text-neutral-300 items-center justify-between py-4 px-4 sm:px-6 lg:px-8">
        <span>Save to apply environment variables.</span>
        <Button loading={loading} disabled={loading} onClick={onSave}>
          Save
        </Button>
      </div>
    </div>
  )
}

export default VariablesInput

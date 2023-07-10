import { TrashIcon } from '@heroicons/react/24/outline'
import classNames from 'classnames'
import { useCallback, useState } from 'react'

import Button from '../components/Button'

type InputProps = {
  label: string
  onChange: (value: string) => void
  value: string
  placeholder: string
  disabled?: boolean
}

const Input = ({
  label,
  onChange,
  value,
  placeholder,
  disabled,
}: InputProps) => {
  return (
    <div className="pr-4">
      <input
        type="text"
        name={label}
        value={value}
        onChange={(e) => {
          onChange(e.target.value)
        }}
        placeholder={placeholder}
        className={classNames(
          'block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-primary-600 sm:text-sm sm:leading-6',
          { 'bg-gray-100': disabled }
        )}
        disabled={disabled}
      />
    </div>
  )
}

type TableInputProps = {
  values: string[][]
  loading: boolean
  cells: number
  labels: string[]
  placeholders: string[]
  onAdd: (value: string[]) => boolean
  onRemove: (i: number) => void
  onSave: () => void
}

const TableInput = (props: TableInputProps) => {
  const { onAdd } = props

  const [current, setCurrent] = useState<string[]>(
    new Array(props.cells).fill('')
  )

  const setCurrentValue = (i: number) => (value: string) => {
    setCurrent((c) => [...c.slice(0, i), value, ...c.slice(i + 1)])
  }
  const onAddHandle: React.FormEventHandler = useCallback(
    (e) => {
      e.preventDefault()

      const value = current.map((c) => c.trim())
      if (value.some((c) => c === '')) {
        return
      }

      const added = onAdd(value)
      if (added) {
        setCurrent((c) => c.map(() => ''))
      }
    },
    [onAdd, current, setCurrent]
  )

  const onRemove = (i: number) => () => {
    props.onRemove(i)
  }

  return (
    <div className="w-full flex flex-col h-full overflow-y-hidden">
      <form onSubmit={onAddHandle}>
        <div className="w-full flex pt-4 px-4 sm:px-6 lg:px-8">
          {current.map((_, i) => (
            <span
              key={i}
              className="flex-1 text-sm grow flex-1 text-left font-semibold text-gray-900"
            >
              {props.labels[i] ?? ''}
            </span>
          ))}
          <span className="w-16"></span>
        </div>

        <div className="w-full flex border-b border-gray-200 py-4 px-4 sm:px-6 lg:px-8">
          {current.map((v, i) => (
            <div className="flex-1 grow">
              <Input
                key={i}
                label={props.labels[i] ?? ''}
                onChange={setCurrentValue(i)}
                value={v}
                placeholder={props.placeholders[i] ?? ''}
                disabled={props.loading}
              />
            </div>
          ))}
          <div className="w-16">
            <Button className="w-full" type="submit" disabled={props.loading}>
              Add
            </Button>
          </div>
        </div>
      </form>

      <div className="h-full overflow-y-scroll">
        {props.values.map((cells, i) => (
          <div className="flex items-center w-full h-16 border-b border-gray-200 px-4 sm:px-6 lg:px-8">
            {cells.map((c, i) => (
              <span
                key={i}
                className={classNames(
                  'flex-1 grow whitespace-nowrap py-4 text-sm font-medium',
                  {
                    'pl-1': i > 0,
                    'text-gray-900': i === 0,
                    'text-gray-500': i > 0,
                  }
                )}
              >
                {c}
              </span>
            ))}
            <span className="whitespace-nowrap py-4 flex items-center justify-center w-16">
              <button onClick={onRemove(i)} disabled={props.loading}>
                <TrashIcon
                  className={classNames('h-5 w-5 flex-shrink-0', {
                    'text-red-600 hover:text-red-400 hover:cursor-pointer':
                      !props.loading,
                    'text-gray-600': props.loading,
                  })}
                  aria-hidden="true"
                />
              </button>
            </span>
          </div>
        ))}
      </div>
      <div className="flex bg-gray-200 items-center justify-between py-4 px-4 sm:px-6 lg:px-8">
        <span>Save to apply environment variables.</span>
        <Button
          loading={props.loading}
          disabled={props.loading}
          onClick={props.onSave}
        >
          Save
        </Button>
      </div>
    </div>
  )
}

export default TableInput

import { TrashIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'

type InputProps = {
  label: string
  onChange: (value: string) => void
  value: string
  placeholder: string
}

const Input = ({ label, onChange, value, placeholder }: InputProps) => {
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
        className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-primary-600 sm:text-sm sm:leading-6"
      />
    </div>
  )
}

type TableInputProps = {
  onSubmit: (content: { name: string; value: string }) => void
}

const TableInput = ({ onSubmit }: TableInputProps) => {
  const [name, setName] = useState('')
  const [value, setValue] = useState('')

  const [envVars, setEnvVars] = useState<
    Array<{ name: string; value: string }>
  >([{ name: 'EXAMPLE', value: 'blabla' }])

  return (
    <div className="w-full flex flex-col">
      <div className="w-full flex pt-4 px-4 sm:px-6 lg:px-8">
        <span className="flex-1 text-sm grow flex-1 text-left font-semibold text-gray-900">
          Name
        </span>
        <span className="flex-1 text-sm grow flex-1 text-left font-semibold text-gray-900">
          Value
        </span>
        <span className="w-16"></span>
      </div>

      <div className="w-full flex border-b border-gray-200 py-4 px-4 sm:px-6 lg:px-8">
        <div className="flex-1 grow">
          <Input
            label="Name"
            onChange={setName}
            value={name}
            placeholder="EXAMPLE_VAR"
          />
        </div>
        <div className="flex-1 grow">
          <Input
            label="Value"
            onChange={setValue}
            value={value}
            placeholder="value123"
          />
        </div>
        <div className="w-16">
          <button
            type="button"
            className="w-full rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
          >
            Add
          </button>
        </div>
      </div>

      {envVars.map((envVar) => (
        <div className="flex w-full">
          <span className="whitespace-nowrap py-4 pl-2 text-sm font-medium text-gray-900 ">
            {envVar.name}
          </span>
          <span className="whitespace-nowrap py-4 pl-2 text-sm text-gray-500">
            {envVar.value}
          </span>
          <span className="whitespace-nowrap py-4 flex items-center justify-center">
            <TrashIcon
              className="h-5 w-5 flex-shrink-0 text-red-600 hover:text-red-400 hover:cursor-pointer"
              aria-hidden="true"
            />
          </span>
        </div>
      ))}
    </div>
  )
}

export default TableInput

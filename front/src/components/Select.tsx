import { Menu } from '@headlessui/react'
import { ArrowPathIcon, ChevronDownIcon } from '@heroicons/react/24/solid'
import React from 'react'

interface Props<T> {
  label?: string
  className?: string
  values: T[]
  value: T
  onChange: (value: T) => void
  children: (value: T) => React.ReactNode
  loading?: boolean
  small?: boolean
  fullWidth?: boolean
  cta?: React.ReactNode
}
function Select<T>(props: Props<T>) {
  const loading = props.loading ?? false

  return (
    <Menu
      as="div"
      className={`relative ${props.fullWidth ? 'w-full' : ''} ${
        props.className ?? ''
      }`}
    >
      {props.label && <label>{props.label}</label>}
      <Menu.Button
        disabled={loading}
        className={`flex items-center justify-between px-4 ${
          props.small ? 'py-2' : 'py-4'
        } border-2 border-outcolor focus:border-gray-400 rounded-md ${
          props.fullWidth ? 'w-full' : ''
        }`}
      >
        {props.children(props.value)}
        {loading ? (
          <ArrowPathIcon className="ml-4 animate-spin h-4 w-4 text-white" />
        ) : (
          <ChevronDownIcon className="ml-4 h-4 w-4 text-primary-500" />
        )}
      </Menu.Button>
      <Menu.Items
        className={`absolute mt-1 bg-black rounded-lg border-outcolor border-2 z-10 ${
          props.fullWidth ? 'w-full' : ''
        }`}
      >
        <div className="bg-dark p-5 rounded-lg text-white font-bold space-y-2">
          {props.values.map((value, i) => (
            <Menu.Item key={i}>
              <button
                className="flex text-sm hover:text-gray-300"
                onClick={() => props.onChange(value)}
              >
                {props.children(value)}
              </button>
            </Menu.Item>
          ))}
          {props.cta && (
            <Menu.Item>
              <div className="flex text-sm hover:text-gray-300">
                {props.cta}
              </div>
            </Menu.Item>
          )}
        </div>
      </Menu.Items>
    </Menu>
  )
}

export default Select

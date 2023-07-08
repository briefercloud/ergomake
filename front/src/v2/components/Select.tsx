import { Listbox, Transition } from '@headlessui/react'
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid'
import { Fragment, useState } from 'react'

import { classNames } from '../utils'

type SelectOptions = {
  options: Array<{ label: string; value: number }>
  onChange: (value: number) => void
}

const Select = ({ options, onChange }: SelectOptions) => {
  const defaultOption = options[0]

  if (!defaultOption) {
    throw new Error('Select component requires at least one option')
  }

  const [selected, setSelected] = useState(defaultOption)

  const actualChangeHandler = (value: number) => {
    setSelected(options.find((option) => option.value === value)!)
    onChange(value)
  }

  return (
    <Listbox value={selected.value} onChange={actualChangeHandler}>
      {({ open }) => (
        <div className="relative w-full bg-primary-400/10 border-r border-primary-400/30">
          <Listbox.Button className="relative w-full cursor-default py-1.5 pl-3 pr-10 text-left text-primary-500 focus:bg-primary-400/20 sm:text-sm sm:leading-6">
            <span className="block truncate">{selected.label}</span>
            <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
              <ChevronUpDownIcon
                className="h-5 w-5 text-gray-400"
                aria-hidden="true"
              />
            </span>
          </Listbox.Button>

          <Transition
            show={open}
            as={Fragment}
            leave="transition ease-in duration-100"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <Listbox.Options className="absolute z-10 max-h-60 w-full overflow-auto bg-white text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
              {options.map((option) => (
                <Listbox.Option
                  key={option.value}
                  className={({ active }) =>
                    classNames(
                      active ? 'bg-primary-600 text-white' : 'text-gray-900',
                      'relative cursor-default select-none py-2 pl-3 pr-9'
                    )
                  }
                  value={option.value}
                >
                  {({ selected, active }) => (
                    <>
                      <span
                        className={classNames(
                          selected ? 'font-semibold' : 'font-normal',
                          'block truncate'
                        )}
                      >
                        {option.label}
                      </span>

                      {selected ? (
                        <span
                          className={classNames(
                            active ? 'text-white' : 'text-indigo-600',
                            'absolute inset-y-0 right-0 flex items-center pr-4'
                          )}
                        >
                          <CheckIcon className="h-5 w-5" aria-hidden="true" />
                        </span>
                      ) : null}
                    </>
                  )}
                </Listbox.Option>
              ))}
            </Listbox.Options>
          </Transition>
        </div>
      )}
    </Listbox>
  )
}

export default Select

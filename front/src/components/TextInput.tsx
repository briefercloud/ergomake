import { MagnifyingGlassIcon } from '@heroicons/react/24/solid'
import React, { useCallback, useState } from 'react'

let id = 0

interface Props {
  label?: string
  placeholder?: string
  value: string
  onChange: (value: string) => void
  Icon?: typeof MagnifyingGlassIcon
  className?: string
  disabled?: boolean
  error?: boolean
}

function TextInput({
  label,
  onChange,
  Icon,
  className,
  placeholder,
  value,
  disabled,
  error,
}: Props) {
  const onChangeHandler: React.ChangeEventHandler<HTMLInputElement> =
    useCallback(
      (e) => {
        onChange(e.target.value)
      },
      [onChange]
    )

  const [inputId] = useState(`input-${id++}`)

  return (
    <div className={className}>
      {label && (
        <label className="text-white text-sm" htmlFor={inputId}>
          {label}
        </label>
      )}
      <div className="relative flex items-center w-full">
        {Icon && <Icon className="w-5 h-5 left-3 absolute" />}
        <input
          id={inputId}
          type="text"
          placeholder={placeholder}
          className={`${Icon ? 'pl-10 ' : 'pl-4 '}rounded border-2 ${
            error ? 'border-red-500' : 'border-outcolor'
          } rounded-md py-4 pr-4 bg-transparent focus:border-gray-400 w-full focus:outline-none`}
          value={value}
          onChange={onChangeHandler}
          disabled={disabled}
        />
      </div>
    </div>
  )
}

export default TextInput

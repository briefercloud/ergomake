import classNames from 'classnames'

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
export default Input
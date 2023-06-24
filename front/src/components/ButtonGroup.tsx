import React, { useCallback } from 'react'

interface ButtonGroupProps {
  value: string
  onChange: (value: string) => void
  children: React.ReactElement<ButtonGroupItemProps>[]
  className?: string
}
function ButtonGroup({
  value,
  onChange,
  children,
  className,
}: ButtonGroupProps) {
  const handleOptionChange = useCallback(
    (optionValue: string) => {
      onChange(optionValue)
    },
    [onChange]
  )

  return (
    <div className={`inline-flex space-x-1 p-1 rounded-md ${className ?? ''}`}>
      {React.Children.map(children, (child) => {
        const optionValue = child.props.value

        return React.cloneElement(child, {
          active: value === optionValue,
          onClick: handleOptionChange,
        })
      })}
    </div>
  )
}

interface ButtonGroupItemProps {
  value: string
  active?: boolean
  onClick?: (value: string) => void
  disabled?: boolean
  children: React.ReactNode
}
function ButtonGroupItem(props: ButtonGroupItemProps) {
  const { onClick, value, active, disabled, children } = props
  const onClickHandler = useCallback(() => {
    onClick?.(value)
  }, [onClick, value])

  return (
    <button
      className={`py-2 px-4 rounded text-sm font-bold${
        active ? ' bg-primary-500' : ''
      }${disabled ? ' text-gray-500' : ' text-white'}`}
      onClick={onClickHandler}
      disabled={disabled}
    >
      {children}
    </button>
  )
}

ButtonGroup.Item = ButtonGroupItem

export default ButtonGroup

import classNames from 'classnames'
import React from 'react'
import colors from 'tailwindcss/colors'

import Loading from './Loading'

type Variant = 'text' | 'contained'
interface Props {
  children: React.ReactNode
  className?: string
  variant?: Variant
  loading?: boolean
  onClick?: () => void
  disabled?: boolean
}
function Button(props: Props) {
  const variant = props.variant ?? 'contained'
  const className = classNames(
    'px-4',
    'py-2',
    'rounded',
    'font-semibold',
    'flex',
    'items-center',
    {
      'bg-primary-500': variant === 'contained' && !props.disabled,
      'bg-gray-500': variant === 'contained' && props.disabled,
      'text-white': !props.disabled,
      'text-gray-400': props.disabled,
    },
    props.className
  )

  let loadingColor = '#f2f2f2'
  if (props.disabled) {
    loadingColor = colors.gray['400']
  }

  return (
    <button
      className={className}
      disabled={props.disabled}
      onClick={props.onClick}
    >
      {props.loading && (
        <Loading className="mr-2" size={5} color={loadingColor} />
      )}
      {props.children}
    </button>
  )
}

export default Button

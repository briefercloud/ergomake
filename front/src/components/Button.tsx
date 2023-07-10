import classNames from 'classnames'

import Loading from './Loading'

interface Props extends React.ButtonHTMLAttributes<unknown> {
  loading?: boolean
}
function Button(props: Props) {
  const className = classNames(
    props.className,
    'rounded-md',
    'px-3',
    'py-2',
    'text-sm',
    'font-semibold',
    'text-white',
    'shadow-sm',
    'focus-visible:outline',
    'focus-visible:outline-2',
    'focus-visible:outline-offset-2',
    'focus-visible:outline-primary-600',
    {
      'bg-primary-600': !props.disabled,
      'bg-gray-400': props.disabled,
      'hover:bg-primary-500': !props.disabled,
      flex: props.loading,
      'items-center': props.loading,
    }
  )

  return (
    <button {...props} className={className}>
      {props.loading && <Loading className="mr-1" size={4} />}
      {props.children}
    </button>
  )
}

export default Button

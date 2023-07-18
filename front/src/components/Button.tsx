import classNames from 'classnames'

import Loading from './Loading'

interface Props extends React.ButtonHTMLAttributes<unknown> {
  loading?: boolean
  tag?: 'button' | 'a'
  href?: string
}

function Button(props: Props) {
  const { href, loading, className, children, ...buttonProps } = props
  const cn = classNames(
    className,
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
    'dark:border dark:border-primary-400',
    {
      'bg-primary-600 dark:bg-primary-800': !buttonProps.disabled,
      'bg-gray-400': buttonProps.disabled,
      'hover:bg-primary-500 dark:hover:bg-primary-700': !buttonProps.disabled,
      flex: loading,
      'items-center': loading,
      'justify-center': loading,
    }
  )

  const Tag = props.tag ?? 'button'

  return (
    <Tag {...buttonProps} className={cn} href={href}>
      {loading && <Loading className="mr-1" size={12} />}
      {children}
    </Tag>
  )
}

export default Button

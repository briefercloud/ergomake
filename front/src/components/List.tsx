import { ChevronRightIcon } from '@heroicons/react/20/solid'
import classNames from 'classnames'
import { useCallback } from 'react'
import { Link } from 'react-router-dom'

type ListItem<T> = {
  name: string
  statusBall?: React.ReactNode
  descriptionLeft: string
  descriptionRight: string
  chevron?: React.ReactNode
  url?: string
  onClick?: (d: T) => void
  data: T
}
function Item<T>(props: ListItem<T>) {
  const { onClick, data } = props
  const onClickHandler = useCallback(() => {
    onClick?.(data)
  }, [onClick, data])

  const item = (
    <>
      <div className="min-w-0 flex-auto space-y-2">
        <div className="flex items-center gap-x-3">
          {props.statusBall}

          <h2 className="min-w-0 text-lg font-semibold leading-6 text-primary-900 dark:text-neutral-200">
            <span className="whitespace-nowrap">{props.name}</span>
          </h2>
        </div>
        <div className="mt-3 flex items-center gap-x-2.5 text-xs leading-5 text-gray-400">
          <p className="truncate">{props.descriptionLeft}</p>
          <svg viewBox="0 0 2 2" className="h-1 w-1 flex-none fill-gray-300">
            <circle cx={1} cy={1} r={1} />
          </svg>
          <p className="whitespace-nowrap">{props.descriptionRight}</p>
        </div>
      </div>
      {props.chevron ?? (
        <ChevronRightIcon
          className="h-5 w-5 flex-none text-gray-400"
          aria-hidden="true"
        />
      )}
    </>
  )

  const className = classNames(
    'relative w-full flex items-center space-x-4 py-4 px-8  border-b border-neutral-200/70 dark:border-neutral-800/70',
    {
      'hover:bg-gray-100 dark:hover:bg-neutral-950/50 hover:cursor-pointer':
        props.url !== undefined,
    },
    { 'hover:cursor-not-allowed': props.url === undefined }
  )

  if (props.url) {
    return (
      <Link className={className} to={props.url}>
        {item}
      </Link>
    )
  }

  if (props.onClick) {
    return (
      <button className={className} onClick={onClickHandler}>
        {item}
      </button>
    )
  }

  return <div className={className}>{item}</div>
}

type ListProps<T> = {
  items: ListItem<T>[]
  emptyState?: React.ReactNode
}
function List<T>({ items, emptyState }: ListProps<T>) {
  if (items.length === 0 && emptyState) {
    return <>{emptyState}</>
  }

  return (
    <ul className="h-full overflow-y-scroll dark:bg-neutral-900">
      {items.map((item, i) => (
        <li key={i}>
          <Item {...item} />
        </li>
      ))}
    </ul>
  )
}

export default List

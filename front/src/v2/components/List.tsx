import { ChevronRightIcon } from '@heroicons/react/20/solid'
import { Link } from 'react-router-dom'

type ListItem = {
  name: string
  statusBall?: React.ReactNode
  descriptionLeft: string
  descriptionRight: string
  url: string
}

type ListProps = {
  items: ListItem[]
}

const List = ({ items }: ListProps) => {
  return (
    <ul className="-mx-4 sm:-mx-6 lg:-mx-8">
      {items.map((item) => (
        <Link to={item.url}>
          <li
            key={item.name}
            className="relative flex items-center space-x-4 py-4 px-8 hover:bg-gray-100 border-b border-gray-200/70 hover:cursor-pointer"
          >
            <div className="min-w-0 flex-auto space-y-2">
              <div className="flex items-center gap-x-3">
                {item.statusBall}

                <h2 className="min-w-0 text-lg font-semibold leading-6 text-primary-900">
                  <span className="whitespace-nowrap">{item.name}</span>
                </h2>
              </div>
              <div className="mt-3 flex items-center gap-x-2.5 text-xs leading-5 text-gray-400">
                <p className="truncate">{item.descriptionLeft}</p>
                <svg
                  viewBox="0 0 2 2"
                  className="h-1 w-1 flex-none fill-gray-300"
                >
                  <circle cx={1} cy={1} r={1} />
                </svg>
                <p className="whitespace-nowrap">{item.descriptionRight}</p>
              </div>
            </div>
            <ChevronRightIcon
              className="h-5 w-5 flex-none text-gray-400"
              aria-hidden="true"
            />
          </li>
        </Link>
      ))}
    </ul>
  )
}

export default List

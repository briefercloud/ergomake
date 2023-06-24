import { ExclamationTriangleIcon } from '@heroicons/react/24/solid'
import React from 'react'
import { Link } from 'react-router-dom'

interface Props {
  owner: string
}
function Alert(props: Props) {
  return (
    <div className="bg-zinc-900 text-gray-300 p-4 rounded-lg flex space-x-2 items-center justify-center">
      <ExclamationTriangleIcon className="w-6 h-6 inline-block text-yellow-500" />
      <span>
        On a free plan you can't have more than three previews at once.{' '}
        <Link
          className="text-primary-500 hover:underline hover:text-primary-400"
          to={`/gh/${props.owner}/purchase`}
        >
          Upgrade to get unlimited previews.
        </Link>
      </span>
    </div>
  )
}

export default Alert

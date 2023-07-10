import { ExclamationTriangleIcon } from '@heroicons/react/24/solid'
import { Link } from 'react-router-dom'

interface Props {
  owner: string
}

const BillingAlert = (props: Props) => {
  return (
    <div className="bg-yellow-100/50 text-gray-300 p-4 border-b border-b-gray-200 flex space-x-2 items-center justify-center font-bold">
      <ExclamationTriangleIcon className="w-6 h-6 inline-block text-yellow-500" />
      <span className="text-gray-600">
        On a free plan you can only have three simultaneous preview links.{' '}
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

export default BillingAlert

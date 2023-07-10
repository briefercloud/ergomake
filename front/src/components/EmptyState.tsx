import { PlusIcon } from '@heroicons/react/20/solid'
import React from 'react'

type EmptyStateProps = {
  title: string
  description: string
  action?: React.ReactNode
  onAction?: () => void
  // Look away
  icon: any
}

const EmptyState = ({
  title,
  description,
  action,
  onAction,
  icon,
}: EmptyStateProps) => {
  const Icon = icon

  const actionComponent =
    action && onAction ? (
      <div className="pt-6">
        <button
          onClick={onAction}
          className="inline-flex items-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600"
        >
          <PlusIcon className="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
          {action}
        </button>
      </div>
    ) : null

  return (
    <div className="px-4 sm:px-6 lg:px-8 pt-6">
      <div className="relative block w-full rounded-lg border-2 border-dashed border-gray-300 p-12 text-center focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">
        <div className="p-4">
          <Icon className="mx-auto h-12 w-12 text-gray-400" />
        </div>
        <h3 className="mt-2 block text-sm font-semibold text-gray-900">
          {title}
        </h3>
        <p className="mt-1 text-sm text-gray-500">{description}</p>

        {actionComponent}
      </div>
    </div>
  )
}

export default EmptyState

import { Dialog, Transition } from '@headlessui/react'
import { IdentificationIcon } from '@heroicons/react/24/outline'
import { Fragment } from 'react'

import Button from './Button'

type PermissionsModalProps = {
  isOpen: boolean
  onClose: () => void
  installationUrl: string
}

const PermissionsModal = ({
  isOpen,
  onClose,
  installationUrl,
}: PermissionsModalProps) => {
  return (
    <Transition.Root show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-500 dark:bg-gray-700 dark:bg-opacity-75 bg-opacity-75 transition-opacity" />
        </Transition.Child>

        <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              enterTo="opacity-100 translate-y-0 sm:scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 translate-y-0 sm:scale-100"
              leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
              <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white dark:bg-neutral-950 px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-md sm:p-6">
                <div>
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-700">
                    <IdentificationIcon
                      className="h-6 w-6 text-green-600 dark:text-green-300"
                      aria-hidden="true"
                    />
                  </div>
                  <div className="mt-3 text-center sm:mt-5">
                    <Dialog.Title
                      as="h3"
                      className="text-base font-semibold leading-6 text-gray-900 dark:text-neutral-200"
                    >
                      We'll ask you for permissions
                    </Dialog.Title>
                    <div className="mt-4 flex gap-y-4 text-left flex-col">
                      <p className="text-sm text-gray-500 dark:text-neutral-400">
                        <span className="font-bold">
                          We need read permissions
                        </span>{' '}
                        so that we can read your{' '}
                        <span className="font-mono">docker-compose.yml</span>{' '}
                        file and build the code that goes into previews.
                      </p>
                      <p className="text-sm text-gray-500 dark:text-neutral-400">
                        <span className="font-bold">
                          We need write permissions
                        </span>{' '}
                        so that we can automatically open a pull-request for you
                        to configure a new repository. We'll <i>never</i>{' '}
                        actually write directly to your repository.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="mt-5 sm:mt-6 flex items-center justify-center text-center">
                  <Button className="w-full" href={installationUrl} tag="a">
                    Add organization
                  </Button>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  )
}

export default PermissionsModal

import { Dialog, Transition } from '@headlessui/react'
import { CheckIcon } from '@heroicons/react/24/outline'
import { Fragment, useCallback, useEffect, useState } from 'react'

import { Repo } from '../hooks/useRepo'
import Button from './Button'

type State = {
  loading: boolean
  error: Error | null
}

interface Props {
  repo: Repo | null
  onClose: (success: boolean) => void
}
function ConfigureRepoModal({ repo, onClose }: Props) {
  const [state, setState] = useState<State>({
    loading: false,
    error: null,
  })
  useEffect(() => {
    setState({ loading: false, error: null })
  }, [repo])

  const handleOnConfigure = useCallback(() => {
    if (!repo) {
      return
    }

    const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/github/owner/${repo.owner}/repos/${repo.name}/configure`
    setState({ loading: true, error: null })

    fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })
      .then(() => {
        setState({ loading: false, error: null })
        onClose(true)
      })
      .catch((err) => {
        setState({ loading: false, error: err })
      })
  }, [repo, setState, onClose])

  const handleOnClose = useCallback(() => {
    if (state.loading) {
      return
    }

    onClose(false)
  }, [state, onClose])

  return (
    <Transition.Root show={repo !== null} as={Fragment}>
      <Dialog as="div" className="relative z-10" onClose={handleOnClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
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
              <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-sm sm:p-6">
                <div>
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
                    <CheckIcon
                      className="h-6 w-6 text-green-600"
                      aria-hidden="true"
                    />
                  </div>
                  <div className="mt-3 text-center sm:mt-5">
                    <Dialog.Title
                      as="h3"
                      className="text-base font-semibold leading-6 text-gray-900"
                    >
                      YOU LIKE BICYCLING?
                    </Dialog.Title>
                    <div className="mt-2">
                      <p className="text-sm text-gray-500">
                        The trail is somewhat wide for singletrack and a bit
                        technical with roots and rocks. Have fun and hang on
                        tight! Amazing views of Castle valley are afforded after
                        50 meters of riding. It is steep. It is not hard to get
                        to and is very fun to ride.
                      </p>
                    </div>
                  </div>
                </div>
                <div className="mt-5 sm:mt-6">
                  <Button
                    disabled={state.loading}
                    loading={state.loading}
                    className="w-full"
                    onClick={handleOnConfigure}
                  >
                    BIKE!
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

export default ConfigureRepoModal

import { Dialog, Transition } from '@headlessui/react'
import {
  ArrowPathIcon,
  CheckIcon,
  FolderPlusIcon,
} from '@heroicons/react/24/outline'
import { Fragment, useCallback, useEffect, useState } from 'react'

import { Repo } from '../hooks/useRepo'
import Button from './Button'

type State = {
  loading: boolean
  error: Error | null
  pullRequestURL: null | string
}

interface Props {
  repo: Repo | null
  onClose: (success: boolean) => void
}

function ConfigureRepoModal({ repo, onClose }: Props) {
  const [state, setState] = useState<State>({
    pullRequestURL: null,
    loading: false,
    error: null,
  })

  useEffect(() => {
    if (repo) {
      setState({
        pullRequestURL: null,
        loading: false,
        error: null,
      })
    }
  }, [repo])

  const handleOnConfigure = useCallback(() => {
    if (!repo) {
      return
    }

    const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/github/owner/${repo.owner}/repos/${repo.name}/configure`
    setState({ loading: true, error: null, pullRequestURL: null })

    fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })
      .then(async (res) => {
        if (!res.ok) {
          throw new Error(res.statusText)
        }

        const response: { pullRequestURL: string } = await res.json()

        setState({
          pullRequestURL: response.pullRequestURL,
          loading: false,
          error: null,
        })
      })
      .catch((err) => {
        setState({ loading: false, error: err, pullRequestURL: null })
      })
  }, [repo, setState])

  const handleOnClose = useCallback(() => {
    if (state.loading) {
      return
    }

    onClose(state.pullRequestURL !== null)
  }, [state, onClose])

  return (
    <Transition.Root show={repo !== null} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={handleOnClose}>
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
              <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white dark:bg-neutral-950 px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-sm sm:p-6">
                {!state.loading && (
                  <div>
                    <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-700">
                      {state.pullRequestURL === null ? (
                        <FolderPlusIcon
                          className="h-6 w-6 text-green-600 dark:text-green-300"
                          aria-hidden="true"
                        />
                      ) : (
                        <CheckIcon
                          className="h-6 w-6 text-green-600 dark:text-green-300"
                          aria-hidden="true"
                        />
                      )}
                    </div>
                    <div className="mt-3 text-center sm:mt-5">
                      <Dialog.Title
                        as="h3"
                        className="text-base font-semibold leading-6 text-gray-900 dark:text-neutral-200"
                      >
                        {!state.pullRequestURL
                          ? 'Automatic configuration'
                          : 'Pull request ready'}
                      </Dialog.Title>
                      <div className="mt-2">
                        {!state.pullRequestURL ? (
                          <p className="text-sm text-gray-500 dark:text-neutral-400">
                            Ergomake will automatically create a pull request
                            with an example configuration file that you can update
                            to configure preview environments.
                          </p>
                        ) : (
                          <>
                            <p className="text-sm text-gray-500 dark:text-neutral-400">
                              Your pull request is ready. Here is a link to view
                              it:
                            </p>
                            <p className="pt-2">
                              <a
                                href={state.pullRequestURL}
                                target="_blank"
                                rel="noreferrer"
                                className="text-primary-500"
                              >
                                {state.pullRequestURL}
                              </a>
                            </p>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                )}
                {state.loading && (
                  <div>
                    <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100 dark:bg-primary-700">
                      <ArrowPathIcon
                        className="h-6 w-6 text-green-600 dark:text-green-300 animate-spin"
                        aria-hidden="true"
                      />
                    </div>
                    <div className="mt-3 text-center sm:mt-5">
                      <Dialog.Title
                        as="h3"
                        className="text-base font-semibold leading-6 text-gray-900 dark:text-neutral-200"
                      >
                        Creating pull request...
                      </Dialog.Title>
                    </div>
                  </div>
                )}

                {!state.loading && !state.pullRequestURL && (
                  <div className="mt-5 sm:mt-6">
                    <Button className="w-full" onClick={handleOnConfigure}>
                      Configure
                    </Button>
                  </div>
                )}
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  )
}

export default ConfigureRepoModal

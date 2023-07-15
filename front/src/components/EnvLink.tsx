import { DocumentDuplicateIcon } from '@heroicons/react/24/outline'
import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid'
import classNames from 'classnames'
import { useEffect, useState } from 'react'
import { CopyToClipboard } from 'react-copy-to-clipboard'

type EnvLinkProps = { link: string }

function EnvLink({ link }: EnvLinkProps) {
  const [isCopied, setIsCopied] = useState(false)
  useEffect(() => {
    if (isCopied) {
      setTimeout(() => {
        setIsCopied(false)
      }, 2000)
    }
  }, [isCopied])

  return (
    <div className="flex rounded-md shadow-sm">
      <CopyToClipboard
        text={link}
        onCopy={() => {
          setIsCopied(true)
        }}
      >
        <div className="group relative flex flex-grow items-stretch focus-within:z-10">
          <div className="absolute inset-y-0 left-0 flex items-center pl-3">
            <DocumentDuplicateIcon
              className={classNames(
                'group-hover:text-primary-400 h-5 w-5 text-gray-400 group-hover:cursor-pointer',
                { 'text-primary-400': isCopied }
              )}
              aria-hidden="true"
            />
          </div>
          <input
            type="text"
            name="link"
            id="link"
            className={classNames(
              'block w-full rounded-none rounded-l-md border-0 py-1.5 pl-10 text-gray-900 ring-1 ring-inset ring-gray-300 dark:ring-neutral-700 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-primary-600 sm:text-sm sm:leading-6 disabled:bg-gray-100 dark:disabled:bg-neutral-800 dark:text-gray-400 hover:text-primary-400 dark:hover:text-primary-400 group-hover:cursor-pointer',
              { 'text-primary-400 dark:text-primary-400': isCopied }
            )}
            value={isCopied ? 'Copied!' : link}
            disabled
          />
        </div>
      </CopyToClipboard>
      <a
        className="group relative -ml-px inline-flex items-center gap-x-1.5 rounded-r-md px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 dark:ring-neutral-700 hover:bg-gray-50 dark:bg-neutral-800 dark:hover:bg-neutral-700"
        target="_blank"
        rel="noreferrer"
        href={link}
      >
        <ArrowTopRightOnSquareIcon
          className="group-hover:text-primary-400 -ml-0.5 h-5 w-5 text-gray-400"
          aria-hidden="true"
        />
      </a>
    </div>
  )
}

export default EnvLink

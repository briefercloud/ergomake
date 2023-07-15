import { ExclamationTriangleIcon } from '@heroicons/react/24/solid'

function AuthAlert() {
  const redirectUrl = `${window.location.protocol}//${window.location.host}`
  const loginUrl = `${process.env.REACT_APP_ERGOMAKE_API}/v2/auth/login?redirectUrl=${redirectUrl}`

  return (
    <div className="bg-yellow-100/50 text-gray-300 p-4 border-b border-b-gray-200 dark:border-t dark:border-yellow-400 flex space-x-2 items-center justify-center font-bold">
      <ExclamationTriangleIcon className="w-6 h-6 inline-block text-yellow-500 dark:text-yellow-400" />
      <span className="text-gray-600 dark:text-gray-300">
        You're no longer authenticated.{' '}
        <a
          className="text-primary-500 hover:underline hover:text-primary-400 dark:text-white"
          href={loginUrl}
        >
          Sign in with GitHub.
        </a>
      </span>
    </div>
  )
}

export default AuthAlert

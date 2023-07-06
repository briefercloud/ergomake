import React from 'react'
import { useLocation } from 'react-router-dom'

import GitHubLogo from '../components/GitHubLogo'
import Logo from '../components/Logo'

const Link = ({ href, children }: { href: string; children: string }) => {
  return (
    <a
      href={href}
      className="font-medium text-primary-600 hover:underline dark:text-primary-500"
    >
      {children}
    </a>
  )
}

const Login = () => {
  const location = useLocation()
  const params = new URLSearchParams(location.search)
  const redirectUrl =
    params.get('redirectUrl') ??
    `${window.location.protocol}//${window.location.host}`
  const loginUrl = `${process.env.REACT_APP_ERGOMAKE_API}/v2/auth/login?redirectUrl=${redirectUrl}`

  return (
    <section className="bg-gray-50 dark:bg-gray-900 relative">
      <div className="flex flex-col items-center justify-center px-6 py-8 mx-auto md:h-screen lg:py-0 relative z-10">
        <div className="w-full bg-white rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">
          <div className="p-6 space-y-4 md:space-y-6 sm:p-8">
            <a href="#" className="flex items-center justify-center p-4">
              <Logo />
            </a>
            <a
              href={loginUrl}
              className="w-full text-white bg-primary-600 hover:bg-primary-700 focus:ring-4 focus:outline-none focus:ring-primary-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800 flex items-center justify-center space-x-2"
            >
              <GitHubLogo />
              <span>Sign in with GitHub</span>
            </a>
            <p className="text-xs font-light text-gray-500 dark:text-gray-400">
              By proceeding, you agree to the{' '}
              <Link href={'https://ergomake.dev/terms-and-conditions/'}>
                Terms of Service
              </Link>{' '}
              and acknowledge you've read the{' '}
              <Link href={'https://ergomake.dev/privacy-policy/'}>
                Privacy Policy
              </Link>
              .
            </p>
          </div>
        </div>
      </div>
    </section>
  )
}

export default Login

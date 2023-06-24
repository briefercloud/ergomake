import React from 'react'
import { useLocation } from 'react-router-dom'

import Background from '../components/Background'
import GitHubLogo from '../components/GitHubLogo'
import Logo from '../components/Logo'
import Pane from '../components/Pane'

function Login() {
  const location = useLocation()
  const params = new URLSearchParams(location.search)
  const redirectUrl = params.get('redirectUrl') ?? `${window.location.protocol}//${window.location.host}`
  const loginUrl = `${process.env.REACT_APP_ERGOMAKE_API}/v2/auth/login?redirectUrl=${redirectUrl}`

  return (
    <Background className="items-center justify-center">
      <Logo />
      <Pane className="flex flex-col space-y-8 mt-16">
        <div>
          <h1 className="text-2xl font-bold pb-1">Hello, world.</h1>
          <p className="text-sm">Login to manage your account.</p>
        </div>

        <a
          href={loginUrl}
          className="flex items-center justify-center bg-primary-500 w-72 font-bold p-4 rounded space-x-2"
        >
          <GitHubLogo />
          <span>Log in with GitHub</span>
        </a>
      </Pane>
    </Background>
  )
}

export default Login

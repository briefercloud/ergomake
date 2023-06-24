import React from 'react'

import Logo from '../components/Logo'

function ErrorLayout() {
  return (
    <div className="absolute inset-0 flex items-center justify-center bg-gray-900">
      <div className="flex flex-col space-y-2 items-center text-white">
        <Logo />
        <p className="text-xl">Oops! Something went wrong.</p>
      </div>
    </div>
  )
}

export default ErrorLayout

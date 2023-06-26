import ArrowPathIcon from '@heroicons/react/24/solid/ArrowPathIcon'
import React, { useEffect, useState } from 'react'

import Logo from '../components/Logo'

interface Props {
  caption?: string
}
function Loading(props: Props) {
  const [showInnerDiv, setShowInnerDiv] = useState(false)
  useEffect(() => {
    const timeout = setTimeout(() => {
      setShowInnerDiv(true)
    }, 1000)

    return () => clearTimeout(timeout)
  }, [])

  return (
    <div className="absolute inset-0 flex items-center justify-center">
      {showInnerDiv && (
        <div className="flex items-center flex-col">
          <div className="flex items-center">
            <Logo />
            <ArrowPathIcon className="ml-4 animate-spin h-8 w-8 text-white" />
          </div>
          {props.caption && <p className="text-white">{props.caption}</p>}
        </div>
      )}
    </div>
  )
}

export default Loading

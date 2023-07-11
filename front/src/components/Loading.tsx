import { ArrowPathIcon } from '@heroicons/react/24/solid'
import React, { useEffect, useState } from 'react'

interface Props {
  size?: number
  className?: string
  color?: string
}
function Loading(props: Props) {
  const color = props.color ?? '#f2f2f2'
  const [showIcon, setShowIcon] = useState(false)
  useEffect(() => {
    const timeout = setTimeout(() => {
      setShowIcon(true)
    }, 200)

    return () => clearTimeout(timeout)
  }, [])

  if (!showIcon) {
    return null
  }

  const size = props.size ?? 8
  const className = props.className ?? ''
  return (
    <ArrowPathIcon
      className={`animate-spin text-white ${className}`}
      fill={color}
      width={size}
      height={size}
    />
  )
}

export default Loading

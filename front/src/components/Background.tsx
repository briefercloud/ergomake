import React from 'react'

type BackgroundProps = {
  children: React.ReactNode
  className?: string
}
const Background = ({ children, className }: BackgroundProps) => {
  return (
    <div
      className={`relative flex flex-col min-h-screen min-w-screen bg-black ${
        className ?? ''
      }`}
    >
      {children}
    </div>
  )
}

export default Background

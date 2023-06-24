import React from 'react'

type PaneProps = {
  children: React.ReactNode
  className?: string
}

const Pane = ({ children, className }: PaneProps) => {
  return (
    <div
      className={`bg-white bg-opacity-5 rounded-2xl p-10 text-white ${className}`}
    >
      {children}
    </div>
  )
}

export default Pane

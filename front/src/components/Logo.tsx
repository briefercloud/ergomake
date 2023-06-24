import React from 'react'

// Returns the logo big by default, if small is true, returns the small logo
const Logo = ({ small }: { small?: boolean }) => {
  if (small) {
    return <LogoSmall />
  }

  return <LogoBig />
}

const LogoSmall = () => {
  return (
    <header className="font-mono font-bold text-2xl tracking-wide">
      <span className="font-mono text-primary-500">e</span>
      <span className="font-mono text-white">m_</span>
    </header>
  )
}

const LogoBig = () => {
  return (
    <header className="font-mono font-bold text-4xl tracking-wide">
      <span className="font-mono text-primary-500">ergo</span>
      <span className="font-mono text-white">make_</span>
    </header>
  )
}

export default Logo

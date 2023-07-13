import React from 'react'

// Returns the logo big by default, if small is true, returns the small logo
const Logo = ({ small }: { small?: boolean }) => {
  let Logo = LogoBig

  if (small) {
    Logo = LogoSmall
  }

  return <Logo />
}

const LogoSmall = () => {
  return (
    <header className="font-mono font-bold text-2xl tracking-wide">
      <span className="font-mono text-primary-400">e</span>
      <span className="font-mono text-dark dark:text-gray-200">m_</span>
    </header>
  )
}

const LogoBig = () => {
  return (
    <header className="font-mono font-bold text-4xl tracking-wide">
      <span className="font-mono text-primary-400">ergo</span>
      <span className="font-mono text-dark dark:text-gray-200">make_</span>
    </header>
  )
}

export default Logo

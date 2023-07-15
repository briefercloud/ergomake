import { createContext, useContext, useState } from 'react'

type UseAuthError = [boolean, (error: boolean) => void]
export const AuthErrorContext = createContext<UseAuthError>([
  false,
  (_: boolean) => {},
])

export const useAuthErrorProvider = (): UseAuthError => {
  return useState(false)
}

export function AuthErrorProvider({ children }: { children: React.ReactNode }) {
  const themeValue = useAuthErrorProvider()

  return (
    <AuthErrorContext.Provider value={themeValue}>
      {children}
    </AuthErrorContext.Provider>
  )
}

export const useAuthError = () => useContext(AuthErrorContext)

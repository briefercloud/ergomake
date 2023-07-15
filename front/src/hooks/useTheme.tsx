import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'

type Theme = 'light' | 'dark'
type UseTheme = [Theme, boolean, () => void]

const loadTheme = (): Theme => {
  const lTheme = localStorage.getItem('theme')
  if (lTheme === 'light') {
    return 'light'
  }

  if (lTheme === 'dark') {
    return 'dark'
  }

  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  if (prefersDark) {
    return 'dark'
  }

  return 'light'
}

const saveTheme = (theme: Theme) => {
  localStorage.setItem('theme', theme)
}

export const ThemeContext = createContext<UseTheme>([
  loadTheme(),
  false,
  () => {},
])

export const useThemeProvider = (): UseTheme => {
  const [theme, setTheme] = useState(loadTheme())
  const [animating, setAnimating] = useState(false)

  useEffect(() => {
    switch (theme) {
      case 'dark':
        document.documentElement.classList.add('dark')
        saveTheme('dark')
        break
      case 'light':
        document.documentElement.classList.remove('dark')
        saveTheme('light')
        break
    }
  }, [theme])

  const toggle = useCallback(() => {
    if (animating) {
      return
    }

    setAnimating(true)
    setTimeout(() => {
      const newTheme = theme === 'dark' ? 'light' : 'dark'
      setTheme(newTheme)
    }, 500)
  }, [theme, animating, setAnimating, setTheme])

  useEffect(() => {
    if (animating) {
      const timer = setTimeout(() => {
        setAnimating(false)
      }, 1000)

      return () => {
        clearTimeout(timer)
      }
    }
  }, [animating])

  return useMemo(() => [theme, animating, toggle], [theme, animating, toggle])
}

export const useTheme = () => useContext(ThemeContext)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const themeValue = useThemeProvider()

  return (
    <ThemeContext.Provider value={themeValue}>{children}</ThemeContext.Provider>
  )
}

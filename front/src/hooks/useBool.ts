import { useCallback, useMemo, useState } from 'react'

type UseBool = [
  boolean,
  {
    setTrue: () => void
    setFalse: () => void
    toggle: () => void
  }
]
const useBool = (initial: boolean): UseBool => {
  const [state, setState] = useState(initial)

  const setTrue = useCallback(() => {
    setState(true)
  }, [setState])

  const setFalse = useCallback(() => {
    setState(false)
  }, [setState])

  const toggle = useCallback(() => {
    setState((s) => !s)
  }, [setState])

  return useMemo(
    () => [
      state,
      {
        setTrue,
        setFalse,
        toggle,
      },
    ],
    [state, setTrue, setFalse, toggle]
  )
}

export default useBool

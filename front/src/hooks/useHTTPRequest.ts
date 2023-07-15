import { useCallback, useEffect, useMemo, useState } from 'react'

import { useAuthError } from './useAuthError'

export type HTTPError =
  | { _tag: 'authentication' }
  | { _tag: 'unexpected'; err: Error }

export type HTTPResponseLoading = { _tag: 'loading' }
export type HTTPResponseSuccess<T> = {
  _tag: 'success'
  body: T
  refreshing: boolean
}
export type HTTPResponseError = { _tag: 'error'; err: HTTPError }
export type HTTPResponse<T> =
  | HTTPResponseLoading
  | HTTPResponseSuccess<T>
  | HTTPResponseError

export const isSuccess = <T>(
  res: HTTPResponse<T>
): res is HTTPResponseSuccess<T> => res._tag === 'success'

export const isLoading = <T>(
  res: HTTPResponse<T>
): res is HTTPResponseLoading => res._tag === 'loading'

export const isError = <T>(res: HTTPResponse<T>): res is HTTPResponseError =>
  res._tag === 'error'

export const map = <A, B>(
  res: HTTPResponse<A>,
  mapper: (a: A) => B
): HTTPResponse<B> => {
  if (isSuccess(res)) {
    return {
      _tag: 'success',
      body: mapper(res.body),
      refreshing: false,
    }
  }

  return res
}

export const orElse = <T>(res: HTTPResponse<T>, fallback: T) => {
  if (isSuccess(res)) {
    return res.body
  }

  return fallback
}

export type FoldCallbacks<A, B> = {
  onError: (res: HTTPResponseError) => B
  onLoading: (res: HTTPResponseLoading) => B
  onSuccess: (res: HTTPResponseSuccess<A>) => B
}
export const fold = <A, B>(
  res: HTTPResponse<A>,
  { onError, onLoading, onSuccess }: FoldCallbacks<A, B>
): B => {
  if (isError(res)) {
    return onError(res)
  }

  if (isLoading(res)) {
    return onLoading(res)
  }

  return onSuccess(res)
}

export const andThen = <A, B>(
  res: HTTPResponse<A>,
  then: (a: HTTPResponseSuccess<A>) => HTTPResponse<B>
): HTTPResponse<B> => {
  if (isSuccess(res)) {
    return then(res)
  }

  return res
}

export type UseHTTPRequest<T> = [HTTPResponse<T>, () => void]
export const useOptionalHTTPRequest = <T>(
  url: string
): UseHTTPRequest<T | null> => {
  const [, setAuthError] = useAuthError()
  const [state, setState] = useState<HTTPResponse<T | null>>({
    _tag: 'loading',
  })

  useEffect(() => {
    const loading = isLoading(state)
    const refreshing = isSuccess(state) && state.refreshing
    if (!loading && !refreshing) {
      return
    }

    const abortController = new AbortController()
    fetch(url, { credentials: 'include', signal: abortController.signal })
      .then(async (res) => {
        if (abortController.signal.aborted) {
          return
        }

        if (res.status === 401 || res.status === 403) {
          setAuthError(true)
          setState({ _tag: 'error', err: { _tag: 'authentication' } })
          return
        }

        if (res.status === 404) {
          setState({
            _tag: 'success',
            body: null,
            refreshing: false,
          })
        }

        if (!res.ok) {
          throw new Error(res.statusText)
        }

        const body: T = await res.json()
        if (abortController.signal.aborted) {
          return
        }

        setState({ _tag: 'success', body, refreshing: false })
      })
      .catch((err) => {
        if (abortController.signal.aborted) {
          return
        }

        setState({ _tag: 'error', err: { _tag: 'unexpected', err } })
      })

    return () => {
      abortController.abort()
    }
  }, [url, state, setAuthError])

  useEffect(() => {
    setState({ _tag: 'loading' })
  }, [url])

  const refetch = useCallback(() => {
    setState((s) => {
      if (!isSuccess(s)) {
        return s
      }

      if (s.refreshing) {
        return s
      }

      return { ...s, refreshing: true }
    })
  }, [setState])

  return useMemo(
    (): UseHTTPRequest<T | null> => [state, refetch],
    [state, refetch]
  )
}

export const useHTTPRequest = <T>(url: string): UseHTTPRequest<T> => {
  const [res, refetch] = useOptionalHTTPRequest<T>(url)

  return useMemo(
    () => [
      andThen(res, ({ body }) => {
        if (body === null) {
          const err = new Error('Not Found')
          err.name = '404'
          return { _tag: 'error', err: { _tag: 'unexpected', err } }
        }

        return { _tag: 'success', body, refreshing: false }
      }),
      refetch,
    ],
    [res, refetch]
  )
}

export type HTTPMutationResponse<T> = HTTPResponse<T> | { _tag: 'pristine' }
export type UseHTTPMutation<P, R = P> = [
  HTTPMutationResponse<R>,
  (data: P) => void
]
export const useHTTPMutation = <P, R = P>(
  url: string
): UseHTTPMutation<P, R> => {
  const [, setAuthError] = useAuthError()
  const [state, setState] = useState<HTTPMutationResponse<R>>({
    _tag: 'pristine',
  })

  const makeRequest = useCallback(
    (data: P) => {
      setState((s) =>
        s._tag === 'success' ? { ...s, refreshing: true } : { _tag: 'loading' }
      )

      fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
        credentials: 'include',
      })
        .then(async (res) => {
          if (res.status === 401 || res.status === 403) {
            setState({ _tag: 'error', err: { _tag: 'authentication' } })
            setAuthError(true)
            return
          }

          if (!res.ok) {
            throw new Error(res.statusText)
          }

          const response: R = await res.json()
          setState({ _tag: 'success', body: response, refreshing: false })
        })
        .catch((err) => {
          setState({ _tag: 'error', err: { _tag: 'unexpected', err } })
        })
    },
    [url, setAuthError]
  )

  return useMemo(() => [state, makeRequest], [state, makeRequest])
}

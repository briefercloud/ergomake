import { useCallback, useEffect, useMemo, useState } from 'react'

export type HTTPError =
  | { _tag: 'authentication' }
  | { _tag: 'unexpected'; err: Error }

export type HTTPResponseLoading = { _tag: 'loading' }
export type HTTPResponseSuccess<T> = {
  _tag: 'success'
  body: T
  loading: boolean
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
      loading: res.loading,
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

const getCacheKey = (url: string) =>
  `ergomake/${process.env.REACT_APP_VERSION}/${url}`

export type UseHTTPRequest<T> = [HTTPResponse<T>, () => void]
export const useOptionalHTTPRequest = <T>(
  url: string,
  skipCache?: boolean
): UseHTTPRequest<T | null> => {
  const [state, setState] = useState<HTTPResponse<T | null>>({
    _tag: 'loading',
  })
  const [cachedData, setCachedData] = useState<
    { _tag: 'miss' } | { _tag: 'hit'; data: T }
  >({ _tag: 'miss' })

  useEffect(() => {
    const cacheKey = getCacheKey(url)
    const cachedData = localStorage.getItem(cacheKey)
    if (cachedData) {
      setCachedData({ _tag: 'hit', data: JSON.parse(cachedData) })
    } else {
      setCachedData({ _tag: 'miss' })
    }
  }, [url])

  useEffect(() => {
    const loading = isLoading(state)
    const refreshing = isSuccess(state) && state.loading
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
          setState({ _tag: 'error', err: { _tag: 'authentication' } })
          localStorage.clear()
          return
        }

        if (res.status === 404) {
          setState({
            _tag: 'success',
            body: null,
            loading: false,
          })
        }

        if (!res.ok) {
          throw new Error(res.statusText)
        }

        const body: T = await res.json()
        if (abortController.signal.aborted) {
          return
        }

        if (!skipCache) {
          localStorage.setItem(getCacheKey(url), JSON.stringify(body))
          setCachedData({ _tag: 'hit', data: body })
        }
        setState({ _tag: 'success', body, loading: false })
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
  }, [skipCache, url, state])

  useEffect(() => {
    setState({ _tag: 'loading' })
  }, [url])

  const refetch = useCallback(() => {
    setState((s) => {
      if (!isSuccess(s)) {
        return s
      }

      if (s.loading) {
        return s
      }

      return { ...s, loading: true }
    })
  }, [setState])

  return useMemo((): UseHTTPRequest<T | null> => {
    if (isLoading(state) && cachedData._tag === 'hit') {
      return [
        {
          _tag: 'success',
          body: cachedData.data,
          loading: true,
        },
        refetch,
      ]
    }

    return [state, refetch]
  }, [cachedData, state, refetch])
}

export const useHTTPRequest = <T>(
  url: string,
  skipCache?: boolean
): UseHTTPRequest<T> => {
  const [res, refetch] = useOptionalHTTPRequest<T>(url, skipCache)

  return useMemo(
    () => [
      andThen(res, ({ body, loading }) => {
        if (body === null) {
          const err = new Error('Not Found')
          err.name = '404'
          return { _tag: 'error', err: { _tag: 'unexpected', err } }
        }

        return { _tag: 'success', body, loading }
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
  const [state, setState] = useState<HTTPMutationResponse<R>>({
    _tag: 'pristine',
  })

  const makeRequest = useCallback(
    (data: P) => {
      setState((s) =>
        s._tag === 'success' ? { ...s, loading: true } : { _tag: 'loading' }
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
            localStorage.clear()
            return
          }

          if (!res.ok) {
            throw new Error(res.statusText)
          }

          const response: R = await res.json()
          setState({ _tag: 'success', body: response, loading: false })
        })
        .catch((err) => {
          setState({ _tag: 'error', err: { _tag: 'unexpected', err } })
        })
    },
    [url]
  )

  return [state, makeRequest]
}

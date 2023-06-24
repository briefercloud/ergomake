import { useCallback, useEffect, useMemo, useState } from 'react'

interface LogEntry {
  timestamp: string
  serviceId: string
  message: string
}

export interface LogData {
  [serviceId: string]: LogEntry[]
}

function useLogs(
  envId: string,
  kind: 'live' | 'build'
): [LogData, Error | null, () => void] {
  const [logData, setLogData] = useState<LogData>({})
  const [error, setError] = useState<Error | null>(null)
  const [eventSource, setEventSource] = useState<EventSource | null>(null)

  useEffect(() => {
    if (!eventSource) {
      return
    }

    if (eventSource.url.includes(envId)) {
      return
    }

    eventSource.close()
    setEventSource(null)
  }, [envId, eventSource])

  useEffect(() => {
    if (!eventSource) {
      setEventSource(
        new EventSource(
          `${process.env.REACT_APP_ERGOMAKE_API}/v2/environments/${envId}/logs/${kind}`,
          { withCredentials: true }
        )
      )
      return
    }

    eventSource.addEventListener('log', (event: MessageEvent) => {
      const logEntry: LogEntry = JSON.parse(event.data)
      setLogData((prevLogData) => {
        const serviceId = logEntry.serviceId
        const updatedLogs = [...(prevLogData[serviceId] || []), logEntry]

        return { ...prevLogData, [serviceId]: updatedLogs }
      })
    })

    eventSource.addEventListener('error', (event: MessageEvent) => {
      setError(new Error(event.data))
      eventSource?.close()
    })
  }, [envId, eventSource, kind])

  const retry = useCallback(() => {
    eventSource?.close()
    setLogData({})
    setError(null)
    setEventSource(null)
  }, [eventSource, setLogData, setError, setEventSource])

  return useMemo(() => [logData, error, retry], [logData, error, retry])
}

export default useLogs

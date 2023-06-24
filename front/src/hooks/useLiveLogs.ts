import useLogs, { LogData } from './useLogs'

function useLiveLogs(envId: string): [LogData, Error | null, () => void] {
  return useLogs(envId, 'live')
}

export default useLiveLogs

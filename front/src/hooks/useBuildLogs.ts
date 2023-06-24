import useLogs, { LogData } from './useLogs'

function useBuildLogs(envId: string): [LogData, Error | null, () => void] {
  return useLogs(envId, 'build')
}

export default useBuildLogs

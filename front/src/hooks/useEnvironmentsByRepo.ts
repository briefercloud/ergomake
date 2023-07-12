import { Environment } from './useEnvironment'
import { UseHTTPRequest, useHTTPRequest } from './useHTTPRequest'

export const useEnvironmentsByRepo = (
  owner: string,
  repo: string
): UseHTTPRequest<Environment[]> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/environments/?owner=${owner}&repo=${repo}`

  return useHTTPRequest<Environment[]>(url)
}

import { HTTPResponse, useHTTPRequest } from './useHTTPRequest'
import { Repo } from './useRepo'

export const useReposByOwner = (owner: string): HTTPResponse<Repo[]> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/github/owner/${owner}/repos`
  return useHTTPRequest<Repo[]>(url)[0]
}

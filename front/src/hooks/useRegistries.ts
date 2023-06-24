import { UseHTTPRequest, useHTTPRequest } from './useHTTPRequest'

export type RegistryProvider = 'ecr' | 'gcr' | 'hub'
export type Registry = {
  id: string
  provider: RegistryProvider
  url: string
}

export const useRegistries = (owner: string): UseHTTPRequest<Registry[]> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/owner/${owner}/registries`
  return useHTTPRequest<Registry[]>(url)
}

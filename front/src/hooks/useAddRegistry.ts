import { UseHTTPMutation, useHTTPMutation } from './useHTTPRequest'
import { Registry, RegistryProvider } from './useRegistries'

export type NewRegistry = {
  provider: RegistryProvider
  url: string
  credentials: string
}

export const useAddRegistry = (
  owner: string
): UseHTTPMutation<NewRegistry, Registry> => {
  const url = `${process.env.REACT_APP_ERGOMAKE_API}/v2/owner/${owner}/registries`

  return useHTTPMutation<NewRegistry, Registry>(url)
}

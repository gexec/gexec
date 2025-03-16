import type { Credential } from '../../../../../client'
import { createContext } from 'radix-vue'
import type { Ref } from 'vue'

export const [useCredentials, provideCredentialsContext] = createContext<{
  credentials: Ref<Credential[]>
  isLoading: Ref<boolean>
  loadCredentials: () => Promise<void>
  addCredentials: (credential: Credential) => void
}>('Credentials')

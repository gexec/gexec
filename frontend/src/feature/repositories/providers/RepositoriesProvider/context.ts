import type { Repository } from '../../../../../client'
import { createContext } from 'radix-vue'
import type { Ref } from 'vue'

export const [useRepositories, provideRepositoriesContext] = createContext<{
  repositories: Ref<Repository[]>
  isLoading: Ref<boolean>
  loadRepositories: () => Promise<void>
  addRepository: (repository: Repository) => void
}>('repositories')

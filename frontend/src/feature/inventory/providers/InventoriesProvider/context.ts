import { createContext } from 'reka-ui'
import type { Inventory } from '../../../../../client'
import type { Ref } from 'vue'

export const [useInventories, provideInventoriesContext] = createContext<{
  inventories: Ref<Inventory[]>
  isLoading: Ref<boolean>
  loadInventories: () => Promise<void>
  addInventory: (inventory: Inventory) => void
}>('Inventories')

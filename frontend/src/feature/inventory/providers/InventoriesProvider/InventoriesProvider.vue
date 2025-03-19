<script setup lang="ts">
import { provideInventoriesContext } from '.'
import { useProjectsStore } from '@/feature/projects/store/projects'
import { ref, unref } from 'vue'
import { listProjectInventories, type Inventory } from '../../../../../client'

const projectStore = useProjectsStore()

const inventories = ref<Inventory[]>([])
const isLoading = ref(false)

async function loadInventories() {
  isLoading.value = true

  const project = unref(projectStore.selectedProject)

  if (!project?.id) {
    throw new Error('No project selected')
  }

  const { data, error } = await listProjectInventories({
    path: { project_id: project.id },
  })

  if (error) {
    isLoading.value = false
    throw error
  }

  inventories.value = data.inventories
  isLoading.value = false
}

function addInventory(inventory: Inventory) {
  inventories.value = [inventory, ...inventories.value]
}

provideInventoriesContext({
  inventories,
  isLoading,
  loadInventories,
  addInventory,
})
</script>

<template>
  <slot />
</template>

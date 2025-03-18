<script setup lang="ts">
import type { Repository } from '../../../../client'
import { ref } from 'vue'
import { provideRepositoriesContext } from './context'
import { useProjectsStore } from '@/feature/projects/store/projects'
import {
  listProjectRepositories,
  type Notification,
} from '../../../../client'

const projectsStore = useProjectsStore()

const repositories = ref<Repository[]>([])
const loadingError = ref<Notification | null>(null)
const isLoading = ref(false)

async function loadRepositories() {
  if (!projectsStore.selectedProject?.id) {
    loadingError.value = {
      status: 400,
      message: 'No project selected',
    }

    return
  }

  isLoading.value = true

  const { data, error } = await listProjectRepositories({
    path: { project_id: projectsStore.selectedProject.id },
  })

  if (error) {
    loadingError.value = error
    isLoading.value = false
    return
  }

  repositories.value = data.repositories
  isLoading.value = false
}

function addRepository(repository: Repository) {
  repositories.value = [repository, ...repositories.value]
}

provideRepositoriesContext({ repositories, loadRepositories, addRepository })
</script>

<template>
  <slot />
</template>

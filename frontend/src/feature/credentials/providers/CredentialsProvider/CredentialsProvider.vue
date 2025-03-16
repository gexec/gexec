<script setup lang="ts">
import { ref, unref } from 'vue'
import { provideCredentialsContext } from '.'
import { type Credential, listProjectCredentials } from '../../../../../client'
import { useProjectsStore } from '@/feature/projects/store/projects'

const projectsStore = useProjectsStore()

const credentials = ref<Credential[]>([])
const isLoading = ref(false)

async function loadCredentials() {
  isLoading.value = true

  const project = unref(projectsStore.selectedProject)

  if (!project?.id) {
    throw new Error('No project selected')
  }

  const { data, error } = await listProjectCredentials({
    path: { project_id: project.id },
  })

  if (error) {
    isLoading.value = false
    throw error
  }

  credentials.value = data.credentials
  isLoading.value = false
}

function addCredentials(credential: Credential) {
  credentials.value = [credential, ...credentials.value]
}

provideCredentialsContext({
  credentials,
  isLoading,
  loadCredentials,
  addCredentials,
})
</script>

<template>
  <slot />
</template>

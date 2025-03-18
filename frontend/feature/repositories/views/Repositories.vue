<script setup lang="ts">
import CredentialsProvider from '@/feature/credentials/providers/CredentialsProvider'
import RepositoriesList from '../components/RepositoriesList'
import RepositoriesProvider from '../providers/RepositoriesProvider'
</script>

<template>
  <CredentialsProvider>
    <RepositoriesProvider>
      <RepositoriesList />
    </RepositoriesProvider>
  </CredentialsProvider>
</template>

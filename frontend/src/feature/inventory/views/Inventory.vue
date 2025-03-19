<script setup lang="ts">
import InventoriesProvider from '../providers/InventoriesProvider'
import InventoriesList from '../components/InventoriesList'
import CredentialsProvider from '@/feature/credentials/providers/CredentialsProvider'
import RepositoriesProvider from '@/feature/repositories/providers/RepositoriesProvider'
</script>

<template>
  <CredentialsProvider>
    <RepositoriesProvider>
      <InventoriesProvider>
        <InventoriesList />
      </InventoriesProvider>
    </RepositoriesProvider>
  </CredentialsProvider>
</template>

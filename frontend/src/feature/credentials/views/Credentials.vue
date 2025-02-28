<script setup lang="ts">
import CredentialsList from '../components/CredentialsList'
import CredentialsProvider from '../providers/CredentialsProvider'
</script>

<template>
  <CredentialsProvider>
    <CredentialsList />
  </CredentialsProvider>
</template>

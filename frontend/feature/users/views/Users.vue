<script setup lang="ts">
import UsersList from '../components/UsersList'
import UsersProvider from '../providers/UsersProvider'
</script>

<template>
  <UsersProvider>
    <UsersList />
  </UsersProvider>
</template>

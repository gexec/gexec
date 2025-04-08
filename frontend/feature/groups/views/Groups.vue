<script setup lang="ts">
import GroupsList from '../components/GroupsList'
import GroupsProvider from '../providers/GroupsProvider'
</script>

<template>
  <GroupsProvider>
    <GroupsList />
  </GroupsProvider>
</template>

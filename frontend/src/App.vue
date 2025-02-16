<script setup lang="ts">
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from '@/components/ui/breadcrumb'
import { Separator } from '@/components/ui/separator'
import { NavigationSidebar } from '@/components/common/navigation-sidebar'
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { RouterView } from 'vue-router'
import { useAuthStore } from './feature/auth/store/auth'
import { Toaster } from './components/ui/toast'
import { onWatcherCleanup, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useProjectsStore } from './feature/projects/store/projects'

const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)

const { loadProjects } = useProjectsStore()

watch(
  isAuthenticated,
  (isAuthenticated) => {
    if (!isAuthenticated) {
      return
    }

    const controller = new AbortController()
    loadProjects({ signal: controller.signal }).catch((error) => {
      console.error(error)
    })

    onWatcherCleanup(() => {
      controller.abort()
    })
  },
  { immediate: true }
)
</script>

<template>
  <SidebarProvider>
    <NavigationSidebar v-if="isAuthenticated" />

    <SidebarInset>
      <header
        v-if="isAuthenticated"
        class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12"
      >
        <div class="flex items-center gap-2 px-4">
          <SidebarTrigger class="-ml-1" />
          <Separator orientation="vertical" class="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>

      <main class="px-4 pt-2 pb-4">
        <RouterView />
      </main>
    </SidebarInset>
  </SidebarProvider>

  <Toaster />
</template>

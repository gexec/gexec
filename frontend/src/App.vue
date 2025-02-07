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

const authStore = useAuthStore()
</script>

<template>
  <SidebarProvider>
    <NavigationSidebar v-if="authStore.isAuthenticated" />

    <SidebarInset>
      <header
        v-if="authStore.isAuthenticated"
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

      <RouterView />
    </SidebarInset>
  </SidebarProvider>

  <Toaster />
</template>

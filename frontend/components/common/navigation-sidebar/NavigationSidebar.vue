<script setup lang="ts">
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from '@/components/ui/sidebar'
import { BadgeCheck, ChevronsUpDown, LogOut, Plus } from 'lucide-vue-next'
import { computed, unref } from 'vue'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { useAuthStore } from '@/feature/auth/store/auth'
import { storeToRefs } from 'pinia'
import { useProjectsStore } from '@/feature/projects/store/projects'
import { getInitials } from '@/lib/utils'
import { RouterLink, useRouter } from 'vue-router'
import { useNavigationLinks } from '@/composables/useNavigationLinks'

const authStore = useAuthStore()
const projectsStore = useProjectsStore()

const { user } = storeToRefs(authStore)
const { projects, selectedProject } = storeToRefs(projectsStore)

const router = useRouter()
const { links } = useNavigationLinks()

const usersInitials = computed(() => getInitials(unref(user).displayName))
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton
                size="lg"
                class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                <div
                  class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"
                >
                  <Avatar
                    class="size-4 shrink-0 bg-gray-900 text-white rounded-lg"
                  >
                    <AvatarFallback>{{
                      getInitials(selectedProject?.name || '')
                    }}</AvatarFallback>
                  </Avatar>
                </div>
                <div class="grid flex-1 text-left text-sm leading-tight">
                  <span class="truncate font-semibold">{{
                    selectedProject?.name
                  }}</span>
                </div>
                <ChevronsUpDown class="ml-auto" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              class="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
              align="start"
              side="bottom"
              :side-offset="4"
            >
              <DropdownMenuLabel class="text-xs text-muted-foreground">
                Projects
              </DropdownMenuLabel>
              <DropdownMenuItem
                v-for="project in projects"
                :key="project.name"
                class="gap-2 p-2"
                @click="
                  router.push({
                    name: unref(router.currentRoute).name,
                    params: {
                      ...unref(router.currentRoute).params,
                      project_slug: project.slug,
                    },
                  })
                "
              >
                <div
                  class="flex size-6 items-center justify-center rounded-sm border"
                >
                  <Avatar class="size-4 shrink-0 bg-transparent">
                    <AvatarFallback>{{
                      getInitials(project.name || '')
                    }}</AvatarFallback>
                  </Avatar>
                </div>
                {{ project.name }}
              </DropdownMenuItem>

              <template v-if="user.isAdmin">
                <DropdownMenuSeparator />
                <DropdownMenuItem class="gap-2 p-2">
                  <div
                    class="flex size-6 items-center justify-center rounded-md border bg-background"
                  >
                    <Plus class="size-4" />
                  </div>
                  <div class="font-medium text-muted-foreground">
                    Add project
                  </div>
                </DropdownMenuItem>
              </template>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Project</SidebarGroupLabel>
        <SidebarMenu>
          <SidebarMenuItem v-for="item in links.project" :key="item.name">
            <SidebarMenuButton as-child>
              <RouterLink :to="item.url">
                <component :is="item.icon" />
                <span>{{ item.name }}</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
      <SidebarGroup v-if="user.isAdmin">
        <SidebarGroupLabel>Admin</SidebarGroupLabel>
        <SidebarMenu>
          <SidebarMenuItem v-for="item in links.admin" :key="item.name">
            <SidebarMenuButton as-child>
              <RouterLink :to="item.url">
                <component :is="item.icon" />
                <span>{{ item.name }}</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton
                size="lg"
                class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                <Avatar class="h-8 w-8 rounded-lg">
                  <AvatarFallback class="rounded-lg">{{
                    usersInitials
                  }}</AvatarFallback>
                </Avatar>
                <div class="grid flex-1 text-left text-sm leading-tight">
                  <span class="truncate font-semibold">{{
                    user.displayName
                  }}</span>
                  <span class="truncate text-xs">{{ user.email }}</span>
                </div>
                <ChevronsUpDown class="ml-auto size-4" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              class="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
              side="bottom"
              align="end"
              :side-offset="4"
            >
              <DropdownMenuLabel class="p-0 font-normal">
                <div
                  class="flex items-center gap-2 px-1 py-1.5 text-left text-sm"
                >
                  <Avatar class="h-8 w-8 rounded-lg">
                    <AvatarFallback class="rounded-lg">{{
                      usersInitials
                    }}</AvatarFallback>
                  </Avatar>
                  <div class="grid flex-1 text-left text-sm leading-tight">
                    <span class="truncate font-semibold">{{
                      user.displayName
                    }}</span>
                    <span class="truncate text-xs">{{ user.email }}</span>
                  </div>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuGroup>
                <DropdownMenuItem>
                  <BadgeCheck />
                  Account
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuItem @select="authStore.signOutUser">
                <LogOut />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail />
  </Sidebar>
</template>

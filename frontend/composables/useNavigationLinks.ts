import { useProjectsStore } from '@/feature/projects/store/projects'
import {
  Archive,
  Book,
  CalendarSync,
  FolderGit2,
  Key,
  LayoutDashboard,
  ListTodo,
  Users,
  Variable,
  Group,
} from 'lucide-vue-next'
import { computed, unref } from 'vue'

export function useNavigationLinks() {
  const projectsStore = useProjectsStore()

  const links = computed(() => {
    const selectedProjectSlug = unref(projectsStore.selectedProject)?.slug || ''

    return {
      project: [
        {
          name: 'Dashboard',
          url: `/${selectedProjectSlug}`,
          icon: LayoutDashboard,
        },
        {
          name: 'Schedules',
          url: '#',
          icon: CalendarSync,
        },
        {
          name: 'Templates',
          url: '#',
          icon: ListTodo,
        },
        {
          name: 'Inventories',
          url: '#',
          icon: Archive,
        },
        {
          name: 'Environments',
          url: '#',
          icon: Variable,
        },
        {
          name: 'Credentials',
          url: `/${selectedProjectSlug}/credentials`,
          icon: Key,
        },
        {
          name: 'Repositories',
          url: `/${selectedProjectSlug}/repositories`,
          icon: FolderGit2,
        },
      ],
      admin: [
        {
          name: 'Projects',
          url: '/admin/projects',
          icon: Book,
        },
        {
          name: 'Users',
          url: '/admin/users',
          icon: Users,
        },
        {
          name: 'Groups',
          url: '/admin/groups',
          icon: Group,
        },
      ],
    }
  })

  return { links }
}

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
          name: 'Templates',
          url: '#',
          icon: ListTodo,
        },
        {
          name: 'Schedule',
          url: '#',
          icon: CalendarSync,
        },
        {
          name: 'Inventory',
          url: `/${selectedProjectSlug}/inventory`,
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
          name: 'Teams',
          url: '#',
          icon: Users,
        },
        {
          name: 'Projects',
          url: '/admin/projects',
          icon: Book,
        },
      ],
    }
  })

  return { links }
}

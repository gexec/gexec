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

export const links = {
  mainNav: [
    {
      name: 'Dashboard',
      url: '#',
      icon: LayoutDashboard,
    },
    {
      name: 'Task Templates',
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
      url: '#',
      icon: Archive,
    },
    {
      name: 'Variable Groups',
      url: '#',
      icon: Variable,
    },
    {
      name: 'Key Stores',
      url: '#',
      icon: Key,
    },
    {
      name: 'Repositories',
      url: '#',
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
      url: '/projects',
      icon: Book,
    },
  ],
}

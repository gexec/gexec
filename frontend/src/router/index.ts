import { useAuthStore } from '@/feature/auth/store/auth'
import NotFound from '@/feature/not-found/views/NotFound.vue'
import { useProjectsStore } from '@/feature/projects/store/projects'
import { unref } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  // Auth
  {
    name: 'SignIn',
    path: '/login',
    component: () => import('../feature/auth/views/SignIn.vue'),
  },

  // Projects
  {
    name: 'Dashboard',
    path: '/:project_slug',
    component: () => import('../feature/dashboard/views/Dashboard.vue'),
    meta: { auth: true },
  },
  {
    name: 'Credentials',
    path: '/:project_slug/credentials',
    component: () => import('../feature/credentials/views/Credentials.vue'),
    meta: { auth: true },
  },
  {
    name: 'Repositories',
    path: '/:project_slug/repositories',
    component: () => import('../feature/repositories/views/Repositories.vue'),
    meta: { auth: true },
  },

  // Admin
  {
    name: 'Projects',
    path: '/admin/projects',
    component: () => import('../feature/projects/views/Projects.vue'),
    meta: { auth: true },
  },

  // 404
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: NotFound },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  const projectsStore = useProjectsStore()

  if (!unref(authStore.isAuthenticated) && to.path === '/') {
    return { name: 'SignIn' }
  }

  if (
    to.meta.auth &&
    !unref(authStore.isAuthenticated) &&
    to.name !== 'SignIn'
  ) {
    return { name: 'SignIn' }
  }

  if (
    unref(authStore.isAuthenticated) &&
    !unref(projectsStore.projects).length
  ) {
    await projectsStore.loadProjects()
  }

  if (
    unref(authStore.isAuthenticated) &&
    Object.hasOwn(to.params, 'project_slug')
  ) {
    const project = unref(projectsStore.projects).find(
      (project) => project.slug === to.params.project_slug
    )

    if (!project) {
      return {
        name: 'NotFound',
        params: { pathMatch: to.path.substring(1).split('/') },
        query: to.query,
        hash: to.hash,
      }
    }

    projectsStore.selectProject(project)
  }

  if (to.path === '/') {
    return {
      name: 'Dashboard',
      params: { project_slug: unref(projectsStore.projects).at(0)!.slug },
    }
  }

  if (to.name === 'SignIn' && unref(authStore.isAuthenticated)) {
    return {
      name: 'Dashboard',
      params: { project_slug: unref(projectsStore.projects).at(0)!.slug },
    }
  }
})

export default router

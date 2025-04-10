import { useAuthStore } from '@/feature/auth/store/auth'
import NotFound from '@/feature/not-found/views/NotFound.vue'
import { useProjectsStore } from '@/feature/projects/store/projects'
import { unref } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    name: 'signin',
    path: '/auth/signin',
    component: () => import('../feature/auth/views/Signin.vue'),
  },
  {
    name: 'callback',
    path: '/auth/callback/:redirect_token',
    component: () => import('../feature/auth/views/Callback.vue'),
  },

  {
    name: 'dashboard',
    path: '/:project_slug',
    component: () => import('../feature/dashboard/views/Dashboard.vue'),
    meta: { auth: true },
  },
  {
    name: 'credentials',
    path: '/:project_slug/credentials',
    component: () => import('../feature/credentials/views/Credentials.vue'),
    meta: { auth: true },
  },
  {
    name: 'repositories',
    path: '/:project_slug/repositories',
    component: () => import('../feature/repositories/views/Repositories.vue'),
    meta: { auth: true },
  },

  {
    name: 'projects',
    path: '/admin/projects',
    component: () => import('../feature/projects/views/Projects.vue'),
    meta: { auth: true },
  },
  {
    name: 'users',
    path: '/admin/users',
    component: () => import('../feature/users/views/Users.vue'),
    meta: { auth: true },
  },
  {
    name: 'groups',
    path: '/admin/groups',
    component: () => import('../feature/groups/views/Groups.vue'),
    meta: { auth: true },
  },

  {
    path: '/:pathMatch(.*)*',
    name: 'notfound',
    component: NotFound,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  const projectsStore = useProjectsStore()

  if (!unref(authStore.isAuthenticated) && to.path === '/') {
    return {
      name: 'signin',
      query: { redirect: to.fullPath },
    }
  }

  if (
    to.meta.auth &&
    !unref(authStore.isAuthenticated) &&
    to.name !== 'signin'
  ) {
    return {
      name: 'signin',
      query: { redirect: to.fullPath },
    }
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
        name: 'notfound',
        params: { pathMatch: to.path.substring(1).split('/') },
        query: to.query,
        hash: to.hash,
      }
    }

    projectsStore.selectProject(project)
  }

  if (to.path === '/') {
    return {
      name: 'dashboard',
      params: { project_slug: unref(projectsStore.projects).at(0)!.slug },
    }
  }

  if (to.name === 'signin' && unref(authStore.isAuthenticated)) {
    return {
      name: 'dashboard',
      params: { project_slug: unref(projectsStore.projects).at(0)!.slug },
    }
  }
})

export default router

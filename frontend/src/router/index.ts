import { useAuthStore } from '@/feature/auth/store/auth'
import { unref } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    name: 'Dashboard',
    path: '/',
    component: () => import('../feature/dashboard/views/Dashboard.vue'),
    meta: { auth: true },
  },

  // Auth
  {
    name: 'SignIn',
    path: '/login',
    component: () => import('../feature/auth/views/SignIn.vue'),
  },

  // Projects
  {
    name: 'Projects',
    path: '/projects',
    component: () => import('../feature/projects/views/Projects.vue'),
    meta: { auth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (
    to.meta.auth &&
    !unref(authStore.isAuthenticated) &&
    to.name !== 'SignIn'
  ) {
    return { name: 'SignIn' }
  }

  if (to.name === 'SignIn' && unref(authStore.isAuthenticated)) {
    return { name: 'Dashboard' }
  }
})

export default router

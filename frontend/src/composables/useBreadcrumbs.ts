import { useProjectsStore } from '@/feature/projects/store/projects'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

export function useBreadcrumbs() {
  const route = useRoute()
  const projectsStore = useProjectsStore()

  const breadcrumbs = computed(() => {
    const res = []

    if (route.params.project_slug) {
      const project = projectsStore.projects.find(
        (p) => p.slug === route.params.project_slug
      )

      if (project) {
        res.push({ name: project.name!, url: `/${project.slug}` })
      }
    }

    res.push({ name: route.name, url: `/${route.params.project_slug}` })

    return res
  })

  return { breadcrumbs }
}

import { listProjects, type Project } from '../../../../client'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useProjectsStore = defineStore('projects', () => {
  const projects = ref<Project[]>([])
  const selectedProject = ref<Project | null>(null)

  async function loadProjects({ signal }: { signal?: AbortSignal } = {}) {
    const { data, error } = await listProjects({ signal })

    if (error) {
      projects.value = []
      throw error
    }

    projects.value = data.projects
  }

  function addProject(project: Project) {
    projects.value = [...projects.value, project]
  }

  function selectProject(project: Project) {
    selectedProject.value = project
  }

  return {
    projects,
    selectedProject,
    loadProjects,
    addProject,
    selectProject,
  }
})

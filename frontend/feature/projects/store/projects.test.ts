import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { Project } from '../../../client'

const mockProjects: Project[] = [
  { id: '1', slug: 'alpha', name: 'Alpha' },
  { id: '2', slug: 'beta', name: 'Beta' },
]

vi.mock('../../../client', () => ({
  listProjects: vi.fn(),
}))

import { listProjects } from '../../../client'
import { useProjectsStore } from './projects'

describe('useProjectsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts with an empty projects list', () => {
    const store = useProjectsStore()
    expect(store.projects).toEqual([])
  })

  it('starts with no selected project', () => {
    const store = useProjectsStore()
    expect(store.selectedProject).toBeNull()
  })

  describe('loadProjects', () => {
    it('populates projects on success', async () => {
      vi.mocked(listProjects).mockResolvedValueOnce({
        data: { projects: mockProjects },
        error: undefined,
        response: new Response(),
      })

      const store = useProjectsStore()
      await store.loadProjects()

      expect(store.projects).toEqual(mockProjects)
    })

    it('resets projects and throws on API error', async () => {
      const apiError = new Error('network error')
      vi.mocked(listProjects).mockResolvedValueOnce({
        data: undefined,
        error: apiError,
        response: new Response(),
      })

      const store = useProjectsStore()
      store.projects = [...mockProjects] // pre-populate

      await expect(store.loadProjects()).rejects.toThrow('network error')
      expect(store.projects).toEqual([])
    })
  })

  describe('addProject', () => {
    it('appends a project to the list', () => {
      const store = useProjectsStore()
      store.projects = [mockProjects[0]]

      store.addProject(mockProjects[1])

      expect(store.projects).toHaveLength(2)
      expect(store.projects[1]).toEqual(mockProjects[1])
    })
  })

  describe('selectProject', () => {
    it('sets the selected project', () => {
      const store = useProjectsStore()
      store.selectProject(mockProjects[0])
      expect(store.selectedProject).toEqual(mockProjects[0])
    })

    it('can switch the selected project', () => {
      const store = useProjectsStore()
      store.selectProject(mockProjects[0])
      store.selectProject(mockProjects[1])
      expect(store.selectedProject?.slug).toBe('beta')
    })
  })
})

import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/feature/projects/store/projects', () => ({
  useProjectsStore: vi.fn(),
}))

import { useProjectsStore } from '@/feature/projects/store/projects'
import { useNavigationLinks } from './useNavigationLinks'

function mountComposable(selectedSlug: string) {
  vi.mocked(useProjectsStore).mockReturnValue({
    selectedProject: { slug: selectedSlug },
  } as ReturnType<typeof useProjectsStore>)

  return useNavigationLinks()
}

describe('useNavigationLinks', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('returns project-scoped navigation links', () => {
    const { links } = mountComposable('my-project')
    const names = links.value.project.map((l) => l.name)
    expect(names).toContain('Dashboard')
    expect(names).toContain('Credentials')
    expect(names).toContain('Repositories')
  })

  it('returns admin navigation links', () => {
    const { links } = mountComposable('my-project')
    const names = links.value.admin.map((l) => l.name)
    expect(names).toContain('Projects')
    expect(names).toContain('Users')
    expect(names).toContain('Groups')
  })

  it('prefixes project links with the selected project slug', () => {
    const { links } = mountComposable('acme')
    const dashboard = links.value.project.find((l) => l.name === 'Dashboard')
    expect(dashboard?.url).toBe('/acme')
  })

  it('uses an empty slug when no project is selected', () => {
    vi.mocked(useProjectsStore).mockReturnValue({
      selectedProject: null,
    } as ReturnType<typeof useProjectsStore>)

    const { links } = useNavigationLinks()
    const dashboard = links.value.project.find((l) => l.name === 'Dashboard')
    expect(dashboard?.url).toBe('/')
  })

  it('includes an icon for every project link', () => {
    const { links } = mountComposable('demo')
    links.value.project.forEach((l) => {
      expect(l.icon).toBeDefined()
    })
  })
})

import { describe, expect, it } from 'vitest'
import { ref } from 'vue'
import { cn, formatSlug, getInitials, valueUpdater } from './utils'

describe('cn', () => {
  it('returns a single class unchanged', () => {
    expect(cn('foo')).toBe('foo')
  })

  it('merges multiple classes', () => {
    expect(cn('px-2', 'py-2')).toBe('px-2 py-2')
  })

  it('deduplicates conflicting Tailwind classes, keeping the last', () => {
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })

  it('ignores falsy values', () => {
    expect(cn('foo', false, undefined, null, 'bar')).toBe('foo bar')
  })
})

describe('getInitials', () => {
  it('returns initials for a full name', () => {
    expect(getInitials('Admin User')).toBe('AU')
  })

  it('returns a single initial for a single word', () => {
    expect(getInitials('Admin')).toBe('A')
  })

  it('handles three-word names', () => {
    expect(getInitials('John Paul Jones')).toBe('JPJ')
  })
})

describe('formatSlug', () => {
  it('converts spaces to hyphens', () => {
    expect(formatSlug('Hello World')).toBe('hello-world')
  })

  it('lowercases the result', () => {
    expect(formatSlug('MyProject')).toBe('myproject')
  })

  it('handles multiple spaces', () => {
    expect(formatSlug('A B C')).toBe('a-b-c')
  })

  it('returns an empty string unchanged', () => {
    expect(formatSlug('')).toBe('')
  })
})

describe('valueUpdater', () => {
  it('sets ref to a direct value', () => {
    const r = ref(0)
    valueUpdater(42, r)
    expect(r.value).toBe(42)
  })

  it('applies an updater function to the current ref value', () => {
    const r = ref(10)
    valueUpdater((prev: number) => prev + 5, r)
    expect(r.value).toBe(15)
  })
})

import { describe, expect, it } from 'vitest'

import { safeInternalRedirect } from '../redirect'

describe('safeInternalRedirect', () => {
  it('keeps internal absolute paths', () => {
    expect(safeInternalRedirect('/dashboard?tab=tasks')).toBe('/dashboard?tab=tasks')
  })

  it('falls back for external or malformed redirects', () => {
    expect(safeInternalRedirect('https://example.com')).toBe('/')
    expect(safeInternalRedirect('//example.com')).toBe('/')
    expect(safeInternalRedirect('javascript:alert(1)')).toBe('/')
    expect(safeInternalRedirect('\\evil')).toBe('/')
  })

  it('uses the provided fallback for empty redirects', () => {
    expect(safeInternalRedirect('', '/home')).toBe('/home')
  })
})

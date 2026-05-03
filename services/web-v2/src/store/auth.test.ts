import { describe, expect, it } from 'vitest'

import { isTokenExpired } from './auth'
import type { AuthTokens } from '#/types/auth'

function tokensWithAccessExpiry(expiresAt: string): AuthTokens {
  return {
    accessToken: { value: 'access-token', expiresAt },
    refreshToken: { value: 'refresh-token', expiresAt: '2099-01-01T00:00:00.000Z' },
  }
}

describe('isTokenExpired', () => {
  it('treats invalid expiry dates as expired', () => {
    expect(isTokenExpired(tokensWithAccessExpiry('not-a-date'))).toBe(true)
  })

  it('detects expired access tokens', () => {
    expect(isTokenExpired(tokensWithAccessExpiry('2000-01-01T00:00:00.000Z'))).toBe(true)
  })

  it('keeps future access tokens valid', () => {
    expect(isTokenExpired(tokensWithAccessExpiry('2099-01-01T00:00:00.000Z'))).toBe(false)
  })
})

import { afterEach, describe, expect, it, vi } from 'vitest'

import { gqlRequest } from '../gql-client'

function stubFetchJson(body: unknown, response?: Partial<Response>) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: vi.fn().mockResolvedValue(body),
      ...response,
    } as unknown as Response),
  )
}

describe('gqlRequest', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('returns a failure when fetch rejects', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    const result = await gqlRequest<{ Ping: string }>('query { Ping }')

    expect(result).toEqual({
      success: false,
      message:
        'Cannot reach GraphQL at http://localhost:8080/query. Start graphql-gateway (e.g. `make dev` / docker compose), set VITE_GRAPHQL_URL for dev/build, or GRAPHQL_URL in the container so config.js exposes it.',
    })
  })

  it('returns the GraphQL error message for non-ok HTTP responses', async () => {
    stubFetchJson(
      { errors: [{ message: 'Unauthorized' }] },
      {
        ok: false,
        status: 401,
      },
    )

    const result = await gqlRequest<{ Me: { id: string } }>('query { Me { id } }')

    expect(result).toEqual({ success: false, message: 'Unauthorized' })
  })

  it('returns the HTTP status for non-ok responses without GraphQL errors', async () => {
    stubFetchJson(
      {},
      {
        ok: false,
        status: 500,
      },
    )

    const result = await gqlRequest<{ Me: { id: string } }>('query { Me { id } }')

    expect(result).toEqual({ success: false, message: 'HTTP 500' })
  })

  it('returns a failure when the server response is not valid JSON', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: vi.fn().mockRejectedValue(new SyntaxError('Unexpected token')),
      } as unknown as Response),
    )

    const result = await gqlRequest<{ Me: { id: string } }>('query { Me { id } }')

    expect(result).toEqual({ success: false, message: 'Invalid server response' })
  })

  it('returns the GraphQL error message for ok responses with errors', async () => {
    stubFetchJson({ errors: [{ message: 'Invalid credentials' }] })

    const result = await gqlRequest<{ Login: { user: { id: string } } }>(
      'mutation { Login { user { id } } }',
    )

    expect(result).toEqual({ success: false, message: 'Invalid credentials' })
  })

  it('returns a failure when the GraphQL response has no data', async () => {
    stubFetchJson({})

    const result = await gqlRequest<{ Me: { id: string } }>('query { Me { id } }')

    expect(result).toEqual({ success: false, message: 'GraphQL response has no data' })
  })

  it('returns data for successful GraphQL responses', async () => {
    const data = { Me: { id: 'user-1' } }
    stubFetchJson({ data })

    const result = await gqlRequest<typeof data>('query { Me { id } }')

    expect(result).toEqual({ success: true, data })
  })
})

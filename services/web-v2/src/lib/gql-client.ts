import { getGraphQLUrl } from '#/lib/graphql-env'

export type GqlSuccess<T> = { success: true; data: T }
export type GqlFailure = {
  success: false
  message: string
  status?: number
  networkError?: boolean
}
export type GqlResult<T> = GqlSuccess<T> | GqlFailure

type GqlRequestOptions = {
  /** If true, sends `Authorization: Bearer` using `getAccessToken()` when provided */
  authenticated?: boolean
  getAccessToken?: () => string | undefined
}

export async function gqlRequest<TData>(
  query: string,
  variables?: Record<string, unknown>,
  options?: GqlRequestOptions,
): Promise<GqlResult<TData>> {
  const url = getGraphQLUrl()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (options?.authenticated) {
    const token = options.getAccessToken?.()
    if (!token) {
      return {
        success: false,
        message: 'Authentication token is missing',
        status: 401,
      }
    }
    if (token) headers.Authorization = `Bearer ${token}`
  }

  let res: Response
  try {
    res = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify({ query, variables }),
    })
  } catch (e) {
    const base = e instanceof Error ? e.message : 'Network error'
    const message =
      /failed to fetch|networkerror|load failed/i.test(base)
        ? `Cannot reach GraphQL at ${url}. Start graphql-gateway (e.g. \`make dev\` / docker compose), set VITE_GRAPHQL_URL for dev/build, or GRAPHQL_URL in the container so config.js exposes it.`
        : base
    return {
      success: false,
      message,
      networkError: true,
    }
  }

  let json: { data?: TData; errors?: { message: string }[] }
  try {
    json = await res.json()
  } catch {
    return { success: false, message: 'Invalid server response' }
  }

  if (!res.ok) {
    const msg = json.errors?.[0]?.message ?? `HTTP ${res.status}`
    return { success: false, message: msg, status: res.status }
  }

  if (json.errors?.length) {
    return { success: false, message: json.errors[0].message }
  }

  if (json.data === undefined) {
    return { success: false, message: 'GraphQL response has no data' }
  }

  return { success: true, data: json.data }
}

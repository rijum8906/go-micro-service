import { getGraphQLUrl } from '#/lib/graphql-env'

export type GqlSuccess<T> = { success: true; data: T }
export type GqlFailure = { success: false; message: string }
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
        ? `Cannot reach GraphQL at ${url}. Start graphql-gateway (e.g. \`make dev\` / docker compose) or set VITE_GRAPHQL_URL.`
        : base
    return {
      success: false,
      message,
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
    return { success: false, message: msg }
  }

  if (json.errors?.length) {
    return { success: false, message: json.errors[0].message }
  }

  if (json.data === undefined) {
    return { success: false, message: 'GraphQL response has no data' }
  }

  return { success: true, data: json.data }
}

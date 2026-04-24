const DEFAULT_GRAPHQL_URL = 'http://localhost:8080/query'

type RelayRuntimeConfig = {
  GRAPHQL_URL?: string
}

function runtimeGraphqlUrlFromWindow(): string | undefined {
  if (typeof window === 'undefined') return undefined
  const cfg = (window as Window & { __CONFIG__?: RelayRuntimeConfig }).__CONFIG__
  const url = cfg?.GRAPHQL_URL
  return typeof url === 'string' && url.length > 0 ? url : undefined
}

/**
 * URL du endpoint GraphQL, ordre de précédence :
 * 1. `window.__CONFIG__.GRAPHQL_URL` (injecté via `public/config.js` ou l’entrypoint Docker)
 * 2. `import.meta.env.VITE_GRAPHQL_URL` (build Vite / dev local)
 * 3. défaut localhost
 */
export function getGraphQLUrl(): string {
  const fromRuntime = runtimeGraphqlUrlFromWindow()
  if (fromRuntime) return fromRuntime

  const fromVite = import.meta.env.VITE_GRAPHQL_URL
  if (typeof fromVite === 'string' && fromVite.length > 0) return fromVite

  return DEFAULT_GRAPHQL_URL
}

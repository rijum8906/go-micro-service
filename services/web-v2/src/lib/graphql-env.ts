const DEFAULT_GRAPHQL_URL = 'http://localhost:8080/query'

export function getGraphQLUrl(): string {
  const url = import.meta.env.VITE_GRAPHQL_URL
  return typeof url === 'string' && url.length > 0 ? url : DEFAULT_GRAPHQL_URL
}

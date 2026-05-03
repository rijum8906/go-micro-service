import { fetchSessionBootstrap } from '#/api/session'
import { isTokenExpired, useAuthStore } from '#/store/auth'

function isAuthFailure(status?: number, message?: string): boolean {
  if (status === 401 || status === 403) return true
  return /\b(unauthorized|unauthenticated|invalid token|expired token)\b/i.test(message ?? '')
}

/**
 * If tokens exist (store / localStorage), validates the session with the gateway.
 * Sets `authReady` to true when finished.
 */
export async function bootstrapAuth(): Promise<void> {
  const store = useAuthStore.getState()
  const token = store.token

  if (!token?.accessToken?.value) {
    store.setAuthReady(true)
    return
  }

  if (isTokenExpired(token)) {
    store.clearAuth()
    store.setAuthReady(true)
    return
  }

  const result = await fetchSessionBootstrap()
  if (result.success) {
    useAuthStore.getState().applySessionBootstrap(result.data)
  } else if (isAuthFailure(result.status, result.message)) {
    useAuthStore.getState().clearAuth()
  }

  useAuthStore.getState().setAuthReady(true)
}

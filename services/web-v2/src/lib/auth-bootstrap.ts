import { fetchSessionBootstrap } from '#/api/session'
import { useAuthStore } from '#/store/auth'

/**
 * If tokens exist (store / localStorage), validates the session with the gateway.
 * Sets `authReady` to true when finished.
 */
export async function bootstrapAuth(): Promise<void> {
  const token = useAuthStore.getState().token
  if (!token?.accessToken?.value) {
    useAuthStore.getState().setAuthReady(true)
    return
  }

  const result = await fetchSessionBootstrap()
  if (result.success) {
    useAuthStore.getState().applySessionBootstrap(result.data)
  } else {
    useAuthStore.getState().clearAuth()
  }

  useAuthStore.getState().setAuthReady(true)
}

import { gqlRequest } from '#/lib/gql-client'
import type { GqlFailure } from '#/lib/gql-client'
import { useAuthStore } from '#/store/auth'
import type { SessionBootstrapData } from '#/types/response'

const SESSION_BOOTSTRAP = `
  query SessionBootstrap {
    GetCurrentSession {
      id
      userId
      deviceId
      ipAddr
      createdAt
      updatedAt
    }
    Me {
      id
      email
      isEmailVerified
      emailVerifiedAt
      twoFactorEnabled
      twoFactorEnabledAt
      createdAt
      updatedAt
    }
    MyProfile {
      id
      userId
      firstName
      lastName
      createdAt
      updatedAt
      avatarUrl
    }
  }
`

export async function fetchSessionBootstrap(): Promise<
  | { success: true; data: SessionBootstrapData }
  | GqlFailure
> {
  const getAccessToken = () => useAuthStore.getState().getAccessTokenValue()

  const result = await gqlRequest<SessionBootstrapData>(
    SESSION_BOOTSTRAP,
    undefined,
    {
      authenticated: true,
      getAccessToken,
    },
  )

  if (!result.success) return result
  return { success: true, data: result.data }
}

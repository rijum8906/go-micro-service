import { gqlRequest } from '#/lib/gql-client'
import { generateDeviceId } from '#/lib/device'
import { useAuthStore } from '#/store/auth'
import type { AuthSuccessPayload } from '#/types/response'
import type { SigninSchemaType, SignupSchemaType } from '#/schemas/auth'

const AUTH_FIELDS = `
  tokens {
    accessToken { value expiresAt }
    refreshToken { value expiresAt }
  }
  user {
    id
    email
    isEmailVerified
    emailVerifiedAt
    twoFactorEnabled
    twoFactorEnabledAt
    createdAt
    updatedAt
  }
  profile {
    id
    userId
    firstName
    lastName
    createdAt
    updatedAt
    avatarUrl
  }
`

const LOGIN_MUTATION = `
  mutation Login($input: LoginInput!) {
    Login(input: $input) {
      ${AUTH_FIELDS}
    }
  }
`

const REGISTER_MUTATION = `
  mutation Register($input: RegisterInput!) {
    Register(input: $input) {
      ${AUTH_FIELDS}
    }
  }
`

const LOGOUT_MUTATION = `
  mutation Logout($input: LogoutInput!) {
    Logout(input: $input) {
      success
      message
    }
  }
`

function meta() {
  return { deviceId: generateDeviceId() }
}

export async function login(
  data: SigninSchemaType,
): Promise<
  | { success: true; data: AuthSuccessPayload }
  | { success: false; message: string }
> {
  const result = await gqlRequest<{ Login: AuthSuccessPayload }>(LOGIN_MUTATION, {
    input: {
      email: data.email,
      password: data.password,
      meta: meta(),
    },
  })

  if (!result.success) return result
  return { success: true, data: result.data.Login }
}

export async function signup(
  data: SignupSchemaType,
): Promise<
  | { success: true; data: AuthSuccessPayload }
  | { success: false; message: string }
> {
  const result = await gqlRequest<{ Register: AuthSuccessPayload }>(
    REGISTER_MUTATION,
    {
      input: {
        email: data.email,
        password: data.password,
        firstName: data.firstName,
        lastName: data.lastName,
        meta: meta(),
      },
    },
  )

  if (!result.success) return result
  return { success: true, data: result.data.Register }
}

export async function logout(): Promise<
  | { success: true; message: string }
  | { success: false; message: string }
> {
  const getAccessToken = () => useAuthStore.getState().getAccessTokenValue()

  const result = await gqlRequest<{ Logout: { success: boolean; message: string } }>(
    LOGOUT_MUTATION,
    { input: { meta: meta() } },
    { authenticated: true, getAccessToken },
  )

  if (!result.success) {
    return result
  }

  const { success, message } = result.data.Logout
  if (!success) {
    return { success: false, message: message || 'Logout refused' }
  }

  useAuthStore.getState().clearAuth()
  return { success: true, message: message || '' }
}

/** Aliases for existing screens */
export const signin = login
export { logout as signout }

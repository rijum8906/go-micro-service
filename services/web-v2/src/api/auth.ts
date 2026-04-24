import type {
  RequestPasswordResetSchemaType,
  SigninSchemaType,
  SignupSchemaType,
} from '#/schemas/auth'
import type { BaseErrorResponse, BaseSuccessResponse } from '#/types/response'
import { useAuthStore } from '#/store/auth'
import { generateDeviceId } from '#/lib/device'
import { gqlRequest } from '#/lib/gql-client'

const getAccessToken = () => useAuthStore.getState().getAccessTokenValue()

const SIGNIN_MUTATION = `
  mutation Login($input: LoginInput!) {
    Login(input: $input) {
      user { id email }
      profile { id userId firstName lastName avatarUrl }
      tokens {
        accessToken { value expiresAt }
        refreshToken { value expiresAt }
      }
    }
  }
`

const SIGNUP_MUTATION = `
  mutation Register($input: RegisterInput!) {
    Register(input: $input) {
      user { id email }
      profile { id userId firstName lastName avatarUrl }
      tokens {
        accessToken { value expiresAt }
        refreshToken { value expiresAt }
      }
    }
  }
`

const SIGNOUT_MUTATION = `
  mutation Logout($input: LogoutInput!) {
    Logout(input: $input) {
      success
      message
    }
  }
`

const REQUEST_PASSWORD_RESET_MUTATION = `
  mutation RequestPasswordReset($input: RequestPasswordResetInput!) {
    RequestPasswordReset(input: $input) {
      success
      message
    }
  }
`

const RESET_PASSWORD_MUTATION = `
  mutation ResetPassword($input: ResetPasswordInput!) {
    ResetPassword(input: $input) {
      success
      message
    }
  }
`

/** GraphQL `Login` / `Register` mutation payload (field names from schema) */
type AuthMutationPayload = {
  user: { id: string; email: string }
  profile: {
    id: string
    userId: string
    firstName: string
    lastName: string
    avatarUrl: string | null
  }
  tokens: {
    accessToken: { value: string; expiresAt: string }
    refreshToken: { value: string; expiresAt: string }
  }
}

export async function signin(
  data: SigninSchemaType,
): Promise<BaseSuccessResponse<{ Login: AuthMutationPayload }> | BaseErrorResponse> {
  const result = await gqlRequest<{ Login: AuthMutationPayload }>(SIGNIN_MUTATION, {
    input: {
      email: data.email,
      password: data.password,
      meta: { deviceId: generateDeviceId() },
    },
  })
  if (!result.success) return { success: false, message: result.message }
  return { success: true, message: '', data: result.data }
}

export async function signup(
  data: SignupSchemaType,
): Promise<BaseSuccessResponse<{ Register: AuthMutationPayload }> | BaseErrorResponse> {
  const result = await gqlRequest<{ Register: AuthMutationPayload }>(SIGNUP_MUTATION, {
    input: {
      email: data.email,
      password: data.password,
      firstName: data.firstName,
      lastName: data.lastName,
      meta: { deviceId: generateDeviceId() },
    },
  })
  if (!result.success) return { success: false, message: result.message }
  return { success: true, message: '', data: result.data }
}

export async function signout(): Promise<BaseSuccessResponse | BaseErrorResponse> {
  const result = await gqlRequest<{ Logout: { success: boolean; message: string } }>(
    SIGNOUT_MUTATION,
    {
      input: { meta: { deviceId: generateDeviceId() } },
    },
    { authenticated: true, getAccessToken },
  )
  if (!result.success) return { success: false, message: result.message }
  return { success: true, message: '', data: result.data }
}

export async function requestPasswordReset(
  data: RequestPasswordResetSchemaType,
): Promise<{ success: boolean; message: string }> {
  const result = await gqlRequest<{
    RequestPasswordReset: { success: boolean; message: string }
  }>(REQUEST_PASSWORD_RESET_MUTATION, {
    input: {
      email: data.email,
      meta: { deviceId: generateDeviceId() },
    },
  })
  if (!result.success) {
    return { success: false, message: result.message }
  }
  const payload = result.data.RequestPasswordReset
  return {
    success: payload.success,
    message: payload.message,
  }
}

export async function resetPassword(input: {
  token: string
  newPassword: string
}): Promise<{ success: boolean; message: string }> {
  const result = await gqlRequest<{
    ResetPassword: { success: boolean; message: string }
  }>(RESET_PASSWORD_MUTATION, {
    input: {
      token: input.token,
      newPassword: input.newPassword,
    },
  })
  if (!result.success) {
    return { success: false, message: result.message }
  }
  const payload = result.data.ResetPassword
  return {
    success: payload.success,
    message: payload.message,
  }
}

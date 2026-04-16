import type {
  RequestPasswordResetSchemaType,
  SigninSchemaType,
  SignupSchemaType,
} from '#/schemas/auth'
import type { AuthResponse, BaseErrorResponse, BaseSuccessResponse } from '#/types/response'
import { useAuthStore } from '#/store/auth'
import { generateDeviceId } from '#/lib/device'

function getGraphQLUrl(): string {
  return (window as any).__CONFIG__?.GRAPHQL_URL ?? 'http://localhost:8080/query'
}

function getAccessToken(): string | undefined {
  return useAuthStore.getState().token?.access_token
}

async function gqlRequest<T>(query: string, variables?: Record<string, unknown>, authenticated = false): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (authenticated) {
    const token = getAccessToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(getGraphQLUrl(), {
    method: 'POST',
    headers,
    body: JSON.stringify({ query, variables }),
  })

  const json = await res.json()

  if (json.errors?.length) {
    return { success: false, message: json.errors[0].message } as T
  }

  return { success: true, data: json.data, message: '' } as T
}

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

export async function signin(data: SigninSchemaType): Promise<AuthResponse | BaseErrorResponse> {
  return gqlRequest(SIGNIN_MUTATION, {
    input: {
      email: data.email,
      password: data.password,
      meta: { deviceId: generateDeviceId() },
    },
  })
}

export async function signup(data: SignupSchemaType): Promise<AuthResponse | BaseErrorResponse> {
  return gqlRequest(SIGNUP_MUTATION, {
    input: {
      email: data.email,
      password: data.password,
      firstName: data.firstName,
      lastName: data.lastName,
      meta: { deviceId: generateDeviceId() },
    },
  })
}

export async function signout(): Promise<BaseSuccessResponse | BaseErrorResponse> {
  return gqlRequest(SIGNOUT_MUTATION, {
    input: { meta: { deviceId: generateDeviceId() } },
  }, true)
}

export async function requestPasswordReset(
  data: RequestPasswordResetSchemaType,
): Promise<{ success: boolean; message: string }> {
  const result = await gqlRequest<{
    success: boolean
    message: string
    data?: {
      RequestPasswordReset: { success: boolean; message: string }
    }
  }>(REQUEST_PASSWORD_RESET_MUTATION, {
    input: {
      email: data.email,
      meta: { deviceId: generateDeviceId() },
    },
  })
  if (!result.success) {
    return { success: false, message: result.message }
  }
  const payload = result.data?.RequestPasswordReset
  return {
    success: payload?.success ?? false,
    message: payload?.message ?? '',
  }
}

export async function resetPassword(input: {
  token: string
  newPassword: string
}): Promise<{ success: boolean; message: string }> {
  const result = await gqlRequest<{
    success: boolean
    message: string
    data?: {
      ResetPassword: { success: boolean; message: string }
    }
  }>(RESET_PASSWORD_MUTATION, {
    input: {
      token: input.token,
      newPassword: input.newPassword,
    },
  })
  if (!result.success) {
    return { success: false, message: result.message }
  }
  const payload = result.data?.ResetPassword
  return {
    success: payload?.success ?? false,
    message: payload?.message ?? '',
  }
}
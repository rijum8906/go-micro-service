import type { SigninSchemaType, SignupSchemaType } from '#/schemas/auth'
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
  mutation Signin($input: SigninInput!) {
    signin(input: $input) {
      account { id email }
      tokens { accessToken refreshToken }
      profiles { id firstName lastName displayName avatarUrl }
    }
  }
`

const SIGNUP_MUTATION = `
  mutation Signup($input: SignupInput!) {
    signup(input: $input) {
      account { id email }
      tokens { accessToken refreshToken }
      profiles { id firstName lastName displayName avatarUrl }
    }
  }
`

const SIGNOUT_MUTATION = `
  mutation Signout($input: SignoutInput!) {
    signout(input: $input) {
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
      metadata: { deviceId: generateDeviceId() },
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
      metadata: { deviceId: generateDeviceId() },
    },
  })
}

export async function signout(): Promise<BaseSuccessResponse | BaseErrorResponse> {
  return gqlRequest(SIGNOUT_MUTATION, {
    input: { metadata: { deviceId: generateDeviceId() } },
  }, true)
}
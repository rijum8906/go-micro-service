import type { SigninSchemaType, SignupSchemaType, ChangePasswordSchemaType } from '#/schemas/auth'
import type { AuthResponse, BaseErrorResponse, BaseSuccessResponse } from '#/types/response'
import { useAuthStore } from '#/store/auth'

function getApiBaseUrl(): string {
  return import.meta.env.VITE_API_BASE_URL ?? ''
}

function getAccessToken(): string | undefined {
  return useAuthStore.getState().token?.access_token
}

async function apiRequest<T>(url: string, options: RequestInit): Promise<T> {
  const res = await fetch(`${getApiBaseUrl()}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  return res.json()
}

async function authenticatedRequest<T>(url: string, options: RequestInit): Promise<T> {
  const accessToken = getAccessToken()
  return apiRequest<T>(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...(options.headers ?? {}),
    },
  })
}

export async function signin(data: SigninSchemaType): Promise<AuthResponse | BaseErrorResponse> {
  return apiRequest('/auth/signin', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function signup(data: SignupSchemaType): Promise<AuthResponse | BaseErrorResponse> {
  return apiRequest('/auth/signup', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function signout(): Promise<BaseSuccessResponse | BaseErrorResponse> {
  return authenticatedRequest('/auth/signout', {
    method: 'POST',
    body: JSON.stringify({}),
  })
}

export async function changePassword(data: ChangePasswordSchemaType): Promise<BaseSuccessResponse | BaseErrorResponse> {
  return authenticatedRequest('/auth/change-password', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}
import { parseBody } from '@/lib/request';
import type {
  ChangePasswordSchemaType,
  CreateProfileSchemaType,
  SigninSchemaType,
  SignupSchemaType,
  UpdateProfileSchemaType,
} from '@/schemas/auth';
import { useAuthStore } from '@/store/auth';
import type {
  AuthResponse,
  BaseErrorResponse,
  BaseSuccessResponse,
  GetProfileResponse,
  UpdateProfileResponse,
} from '@/types/response';

function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? '';
}

function getAccessToken(): string | undefined {
  return useAuthStore.getState().token?.access_token;
}

async function apiRequest<T>(
  url: string,
  options: RequestInit,
): Promise<T> {
  const res = await fetch(`${getApiBaseUrl()}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  return res.json();
}

async function authenticatedRequest<T>(
  url: string,
  options: RequestInit,
): Promise<T> {
  const accessToken = getAccessToken();
  return apiRequest<T>(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...(options.headers ?? {}),
    },
  });
}

export async function signin(
  data: SigninSchemaType,
): Promise<AuthResponse | BaseErrorResponse> {
  return apiRequest('/auth/signin', {
    method: 'POST',
    body: JSON.stringify(parseBody(data)),
  });
}

export async function signup(
  data: SignupSchemaType,
): Promise<AuthResponse | BaseErrorResponse> {
  return apiRequest('/auth/signup', {
    method: 'POST',
    body: JSON.stringify(parseBody(data)),
  });
}

export async function signout(): Promise<
  BaseSuccessResponse | BaseErrorResponse
> {
  return authenticatedRequest('/auth/signout', {
    method: 'POST',
    body: JSON.stringify(parseBody({})),
  });
}

export async function changePassword(
  data: ChangePasswordSchemaType,
): Promise<BaseSuccessResponse | BaseErrorResponse> {
  return authenticatedRequest('/auth/change-password', {
    method: 'POST',
    body: JSON.stringify(parseBody(data)),
  });
}

export async function getProfile(
  id: string,
): Promise<GetProfileResponse | BaseErrorResponse> {
  return authenticatedRequest(`/profiles/${id}`, {
    method: 'GET',
  });
}

export async function createProfile(
  data: CreateProfileSchemaType,
): Promise<AuthResponse | BaseErrorResponse> {
  const formData = new FormData();
  formData.append('firstName', data.firstName);
  formData.append('lastName', data.lastName);
  if (data.displayName) formData.append('displayName', data.displayName);
  if (data.avatar) formData.append('avatar', data.avatar);

  const accessToken = getAccessToken();
  const res = await fetch(`${getApiBaseUrl()}/profiles`, {
    method: 'POST',
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
    body: formData,
  });
  return res.json();
}

export async function updateProfile(
  id: string,
  data: UpdateProfileSchemaType,
): Promise<UpdateProfileResponse | BaseErrorResponse> {
  const formData = new FormData();
  if (data.firstName) formData.append('firstName', data.firstName);
  if (data.lastName) formData.append('lastName', data.lastName);
  if (data.displayName) formData.append('displayName', data.displayName);
  if (data.avatar) formData.append('avatar', data.avatar);

  const accessToken = getAccessToken();
  const res = await fetch(`${getApiBaseUrl()}/profiles/${id}`, {
    method: 'PUT',
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
    body: formData,
  });
  return res.json();
}

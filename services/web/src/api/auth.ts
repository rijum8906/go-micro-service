import type {
  AuthResponse,
  BaseErrorResponse,
  BaseSuccessResponse,
  ErrorResponse,
  GetProfileResponse,
} from '@/types/response';
import { api } from './axios';
import { parseBody } from '@/lib/request';
import type { AxiosError } from 'axios';
import type {
  CreateProfileSchemaType,
  SigninSchemaType,
  SignupSchemaType,
  UpdateProfileSchemaType,
} from '@/schemas/auth';

export async function signin(data: SigninSchemaType) {
  try {
    const body = parseBody<SigninSchemaType>(data);
    const res = await api.post<AuthResponse>('/auth/signin', body);

    if (res.data.data) {
      return res.data;
    }

    return res.data as BaseErrorResponse;
  } catch (error) {
    const err = error as AxiosError<BaseErrorResponse>;
    return err.response?.data as BaseErrorResponse;
  }
}

export async function signup(data: SignupSchemaType) {
  try {
    const body = parseBody<SignupSchemaType>(data);
    const res = await api.post<AuthResponse>('/auth/signup', body);
    if (res.data.data) {
      return res.data;
    }
    return res.data as BaseErrorResponse;
  } catch (error) {
    const err = error as AxiosError<BaseErrorResponse>;
    return err.response?.data as BaseErrorResponse;
  }
}

export async function signout() {
  try {
    const res = await api.post('/auth/signout');
    if (res.data.success) {
      return res.data;
    }

    return res.data as BaseErrorResponse;
  } catch (error) {
    const err = error as AxiosError<BaseErrorResponse>;
    return err.response?.data as BaseErrorResponse;
  }
}

// Profile
export async function getProfile(id: string) {
  try {
    const res = await api.get<GetProfileResponse>(`/profiles/${id}`);
    if (res.data.success) {
      return res.data;
    }

    return res.data as ErrorResponse;
  } catch (error) {
    const err = error as AxiosError<ErrorResponse>;
    return err.response?.data as ErrorResponse;
  }
}

export async function updateProfile(id: string, data: FormData) {
  try {
    const res = await api.put<GetProfileResponse>(`/profiles/${id}`, data, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    return res.data;
  } catch (error) {
    const err = error as AxiosError<BaseErrorResponse>;
    return err.response?.data as BaseErrorResponse;
  }
}

export async function deleteProfile(id: string) {
  try {
    const res = await api.delete(`/profiles/${id}`);
    if (res.data.success) {
      return res.data as BaseSuccessResponse;
    }

    return res.data as BaseErrorResponse;
  } catch (error) {
    const err = error as AxiosError<ErrorResponse>;
    return err.response?.data as BaseErrorResponse;
  }
}

export async function createProfile(data: CreateProfileSchemaType) {
  try {
    const res = await api.post<GetProfileResponse>(
      `/profiles`,
      parseBody(data),
    );
    if (res.data.success) {
      return res.data;
    }

    return res.data as ErrorResponse;
  } catch (error) {
    const err = error as AxiosError<ErrorResponse>;
    return err.response?.data as ErrorResponse;
  }
}

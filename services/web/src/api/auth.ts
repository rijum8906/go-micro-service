import { createServerFn } from '@tanstack/react-start';

import { parseBody } from '@/lib/request';
import { apiAdapter } from '@/lib/server-api';
import type {
  ChangePasswordSchemaType,
  CreateProfileSchemaType,
  SigninSchemaType,
  SignupSchemaType,
} from '@/schemas/auth';
import { useAuthStore } from '@/store/auth';
import type {
  AuthResponse,
  BaseErrorResponse,
  BaseSuccessResponse,
  ErrorResponse,
  GetProfileResponse,
} from '@/types/response';

type AuthenticatedInput = {
  accessToken: string;
};

type AuthPayload<T> = T & {
  metadata: unknown;
};

const TOKEN_FORM_FIELD = '_accessToken';
const PROFILE_ID_FORM_FIELD = '_profileId';

function getAccessToken() {
  return useAuthStore.getState().token?.access_token;
}

function unauthorizedResponse(): BaseErrorResponse {
  return {
    success: false,
    message: 'Authentication required',
  };
}

function appendFormValue(formData: FormData, key: string, value: unknown) {
  if (value === undefined || value === null) return;
  formData.append(key, String(value));
}

function readFormField(formData: FormData, key: string) {
  const value = formData.get(key);
  return typeof value === 'string' ? value : null;
}

const signinServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: unknown) => data)
  .handler(async ({ data }) =>
    apiAdapter<AuthResponse, BaseErrorResponse>({
      method: 'POST',
      url: '/auth/signin',
      data: data as AuthPayload<SigninSchemaType>,
    }),
  );

const signupServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: unknown) => data)
  .handler(async ({ data }) =>
    apiAdapter<AuthResponse, BaseErrorResponse>({
      method: 'POST',
      url: '/auth/signup',
      data: data as AuthPayload<SignupSchemaType>,
    }),
  );

const signoutServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: unknown) => data)
  .handler(async ({ data }) => {
    const payload = data as AuthenticatedInput;

    return apiAdapter<
      BaseSuccessResponse<Record<string, never>>,
      BaseErrorResponse
    >({
      method: 'POST',
      url: '/auth/signout',
      accessToken: payload.accessToken,
    });
  });

const getProfileServerFn = createServerFn({ method: 'GET' })
  .inputValidator((data: unknown) => data)
  .handler(async ({ data }) => {
    const payload = data as { id: string } & AuthenticatedInput;

    return apiAdapter<GetProfileResponse, ErrorResponse>({
      method: 'GET',
      url: `/profiles/${payload.id}`,
      accessToken: payload.accessToken,
    });
  });

const updateProfileServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: FormData) => data)
  .handler(async ({ data }) => {
    const formData = data;
    const accessToken = readFormField(formData, TOKEN_FORM_FIELD);
    const profileId = readFormField(formData, PROFILE_ID_FORM_FIELD);

    if (!accessToken || !profileId) {
      return unauthorizedResponse();
    }

    formData.delete(TOKEN_FORM_FIELD);
    formData.delete(PROFILE_ID_FORM_FIELD);

    return apiAdapter<GetProfileResponse, BaseErrorResponse>({
      method: 'PUT',
      url: `/profiles/${profileId}`,
      data: formData,
      accessToken,
    });
  });

const deleteProfileServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: unknown) => data)
  .handler(async ({ data }) => {
    const payload = data as { id: string } & AuthenticatedInput;

    return apiAdapter<
      BaseSuccessResponse<Record<string, never>>,
      BaseErrorResponse
    >({
      method: 'DELETE',
      url: `/profiles/${payload.id}`,
      accessToken: payload.accessToken,
    });
  });

const createProfileServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: FormData) => data)
  .handler(async ({ data }) => {
    const formData = data;
    const accessToken = readFormField(formData, TOKEN_FORM_FIELD);

    if (!accessToken) {
      return unauthorizedResponse();
    }

    formData.delete(TOKEN_FORM_FIELD);

    return apiAdapter<GetProfileResponse, ErrorResponse>({
      method: 'POST',
      url: '/profiles',
      data: formData,
      accessToken,
    });
  });

const changePasswordServerFn = createServerFn({ method: 'POST' })
  .inputValidator((data: unknown) => data)
  .handler(async ({ data }) => {
    const payload = data as {
      body: AuthPayload<ChangePasswordSchemaType>;
    } & AuthenticatedInput;

    return apiAdapter<
      BaseSuccessResponse<Record<string, never>>,
      BaseErrorResponse
    >({
      method: 'PUT',
      url: '/auth/password',
      data: payload.body,
      accessToken: payload.accessToken,
    });
  });

export async function signin(
  data: SigninSchemaType,
): Promise<AuthResponse | BaseErrorResponse> {
  return signinServerFn({
    data: parseBody(data),
  });
}

export async function signup(
  data: SignupSchemaType,
): Promise<AuthResponse | BaseErrorResponse> {
  return signupServerFn({
    data: parseBody(data),
  });
}

export async function signout(): Promise<
  BaseSuccessResponse<Record<string, never>> | BaseErrorResponse
> {
  const accessToken = getAccessToken();

  if (!accessToken) {
    return unauthorizedResponse();
  }

  return signoutServerFn({
    data: { accessToken },
  });
}

export async function getProfile(
  id: string,
): Promise<GetProfileResponse | ErrorResponse | BaseErrorResponse> {
  const accessToken = getAccessToken();

  if (!accessToken) {
    return unauthorizedResponse();
  }

  return getProfileServerFn({
    data: { id, accessToken },
  });
}

export async function updateProfile(
  id: string,
  data: FormData,
): Promise<GetProfileResponse | BaseErrorResponse> {
  const accessToken = getAccessToken();

  if (!accessToken) {
    return unauthorizedResponse();
  }

  const formData = new FormData();

  for (const [key, value] of data.entries()) {
    formData.append(key, value);
  }

  formData.append(PROFILE_ID_FORM_FIELD, id);
  formData.append(TOKEN_FORM_FIELD, accessToken);

  return updateProfileServerFn({
    data: formData,
  });
}

export async function deleteProfile(
  id: string,
): Promise<BaseSuccessResponse<Record<string, never>> | BaseErrorResponse> {
  const accessToken = getAccessToken();

  if (!accessToken) {
    return unauthorizedResponse();
  }

  return deleteProfileServerFn({
    data: { id, accessToken },
  });
}

export async function createProfile(
  data: CreateProfileSchemaType,
): Promise<GetProfileResponse | ErrorResponse | BaseErrorResponse> {
  const accessToken = getAccessToken();

  if (!accessToken) {
    return unauthorizedResponse();
  }

  const { avatar, ...rest } = data;
  const profile = parseBody(rest);
  const formData = new FormData();

  appendFormValue(formData, 'firstName', profile.firstName);
  appendFormValue(formData, 'lastName', profile.lastName);
  appendFormValue(formData, 'displayName', profile.displayName);
  formData.append('metadata', JSON.stringify(profile.metadata));

  if (avatar) {
    formData.append('avatar', avatar);
  }

  formData.append(TOKEN_FORM_FIELD, accessToken);

  return createProfileServerFn({
    data: formData,
  });
}

export async function changePassword(
  data: ChangePasswordSchemaType,
): Promise<BaseSuccessResponse<Record<string, never>> | BaseErrorResponse> {
  const accessToken = getAccessToken();

  if (!accessToken) {
    return unauthorizedResponse();
  }

  return changePasswordServerFn({
    data: {
      body: parseBody(data),
      accessToken,
    },
  });
}

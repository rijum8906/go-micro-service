import type { GqlProfileFields, BaseErrorResponse, BaseSuccessResponse } from '#/types/response'
import { generateDeviceId } from '#/lib/device'
import { gqlRequest } from '#/lib/gql-client'

const UPDATE_PROFILE_NAME = `
  mutation UpdateProfileName($input: UpdateProfileNameInput!) {
    UpdateProfileName(input: $input) {
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

const GENERATE_SCOPED_TOKEN = `
  mutation GenerateScopedToken($input: GenerateScopedTokenInput!) {
    GenerateScopedToken(input: $input) {
      token {
        value
        expiresAt
      }
    }
  }
`

const CHANGE_PASSWORD = `
  mutation ChangePassword($input: ChangePasswordInput!) {
    ChangePassword(input: $input) {
      success
      message
    }
  }
`

interface ScopedTokenResponse {
  GenerateScopedToken: {
    token: {
      value: string
      expiresAt: string
    }
  }
}

export async function updateProfileName(
  params: { profileId: string; firstName: string; lastName: string },
  getAccessToken: () => string | undefined,
): Promise<BaseSuccessResponse<{ UpdateProfileName: GqlProfileFields }> | BaseErrorResponse> {
  const result = await gqlRequest<{ UpdateProfileName: GqlProfileFields }>(
    UPDATE_PROFILE_NAME,
    {
      input: {
        profileId: params.profileId,
        firstName: params.firstName,
        lastName: params.lastName,
      },
    },
    { authenticated: true, getAccessToken },
  )
  if (!result.success) return { success: false, message: result.message }
  return { success: true, message: '', data: result.data }
}

export async function generateScopedToken(
  params: { scope: string; authMethod: string; authValue: string },
  getAccessToken: () => string | undefined,
): Promise<BaseSuccessResponse<ScopedTokenResponse> | BaseErrorResponse> {
  const result = await gqlRequest<ScopedTokenResponse>(
    GENERATE_SCOPED_TOKEN,
    {
      input: {
        scope: params.scope,
        authMethod: params.authMethod,
        authValue: params.authValue,
        meta: { deviceId: generateDeviceId() },
      },
    },
    { authenticated: true, getAccessToken },
  )
  if (!result.success) return { success: false, message: result.message }
  return { success: true, message: '', data: result.data }
}

export async function changePasswordMutation(
  params: { token: string; newPassword: string },
  getAccessToken: () => string | undefined,
): Promise<BaseSuccessResponse<{ ChangePassword: { success: boolean; message: string } }> | BaseErrorResponse> {
  const result = await gqlRequest<{ ChangePassword: { success: boolean; message: string } }>(
    CHANGE_PASSWORD,
    {
      input: {
        token: params.token,
        newPassword: params.newPassword,
        meta: { deviceId: generateDeviceId() },
      },
    },
    { authenticated: true, getAccessToken },
  )
  if (!result.success) return { success: false, message: result.message }
  return { success: true, message: '', data: result.data }
}
import type { AuthTokens } from './auth'
import type { Profile, Session, User } from './models'
import type { GetProfileResult } from './result'

/** `Profile` fields as returned by the gateway (no client-derived fields) */
export type GqlProfileFields = {
  id: string
  userId: string
  firstName: string
  lastName: string
  createdAt: string
  updatedAt: string
  avatarUrl: string | null
}

export interface ErrorDetail {
  field: string
  message: string
}

export interface BaseResponse {
  success: boolean
  message: string
}

export interface BaseSuccessResponse<T = unknown> {
  success: true
  message: string
  data?: T
}

export interface BaseErrorResponse {
  success: false
  message: string
  errors?: ErrorDetail[]
}

export interface ErrorResponse extends BaseResponse {
  errors?: ErrorDetail[]
}

/** Payload after Login / Register (client-facing fields) */
export interface AuthSuccessPayload {
  tokens: AuthTokens
  user: User
  profile: GqlProfileFields
}

export interface UpdateProfileResponse extends BaseResponse {
  data: Profile
}

export interface GetProfileResponse extends BaseResponse {
  data: GetProfileResult
}

/** Raw `SessionBootstrap` query data (GraphQL field names) */
export interface SessionBootstrapData {
  GetCurrentSession: Session
  Me: User
  MyProfile: GqlProfileFields
}

import type { Token } from './auth';
import type { Account, Profile } from './models';
import type { GetProfileResult } from './result';

export interface ErrorDetail {
  field: string;
  message: string;
}

export interface Metadata {
  deviceId: string;
}

export interface BaseResponse {
  success: boolean;
  message: string;
}

export interface BaseSuccessResponse<T = unknown> {
  success: true;
  message: string;
  data?: T;
}

export interface BaseErrorResponse {
  success: false;
  message: string;
  errors?: ErrorDetail[];
}

export interface ErrorResponse extends BaseResponse {
  errors?: ErrorDetail[];
}

export interface AuthResponse extends BaseResponse {
  data: {
    account: Account;
    profiles: Profile[];
    token: Token;
  };
}

export interface UpdateProfileResponse extends BaseResponse {
  data: Profile;
}

export interface GetProfileResponse extends BaseResponse {
  data: GetProfileResult;
}

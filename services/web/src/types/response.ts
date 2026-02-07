// @/types/response.ts
import { Account, Profile, Token } from "./auth";

export interface ErrorDetail {
  field: string;
  message: string;
}

export interface BaseResponse {
  success: boolean;
  message: string;
}

// Matches your Go BaseErrorResponse
export interface ErrorResponse extends BaseResponse {
  errors?: ErrorDetail[];
}

// Matches your Go BaseSuccessResponse[*dto.AuthResponse]
export interface AuthResponse extends BaseResponse {
  data: {
    account: Account;
    profiles: Profile[];
    token: Token;
  };
}

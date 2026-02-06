import { Account, Token } from "./auth";

export interface Response {
  success: boolean;
  message: string;
  errors?: ErrorResponse[];
}

export interface ErrorResponse {
  field: string;
  message: string;
}

export interface AuthResponse extends Response {
  data: {
    account: Account
    token: Token;
  }
}

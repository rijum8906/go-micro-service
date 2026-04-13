/** GraphQL `Token` scalar shape */
export interface GqlToken {
  value: string
  expiresAt: string
}

/** Gateway `AuthTokens` type */
export interface AuthTokens {
  accessToken: GqlToken
  refreshToken: GqlToken
}

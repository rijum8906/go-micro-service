# GraphQL Gateway API

This document provides a simple, professional reference for the GraphQL API exposed by the gateway service.

## Endpoint

- GraphQL endpoint: `/query`
- Playground: `/` in development mode only

## Authentication

Operations that read or change authenticated user data require an access token in the `Authorization` header.

```http
Authorization: Bearer <access_token>
```

Several mutations and inputs also require request metadata:

```graphql
input RequestMetaInput {
  deviceId: String!
}
```

`deviceId` identifies the client device making the request.

## Scalars

```graphql
scalar DateTime
```

`DateTime` is used for timestamps such as creation, update, and expiry values.

## Queries

### `GetSessions`

Returns a paginated list of sessions for the authenticated user.

```graphql
GetSessions(input: GetSessionsInput): [Session!]
```

Arguments:

```graphql
input GetSessionsInput {
  paginationRequest: PaginationInput!
}

input PaginationInput {
  page: Int!
  limit: Int!
}
```

### `GetActiveSessions`

Returns active sessions for the authenticated user. A scoped token is required by the schema.

```graphql
GetActiveSessions(input: ScopedTokenInput!): [Session!]
```

Arguments:

```graphql
input ScopedTokenInput {
  scopedToken: String!
  meta: RequestMetaInput!
}
```

### `GetCurrentSession`

Returns the current authenticated session.

```graphql
GetCurrentSession: Session!
```

### `MyProfile`

Returns the profile of the authenticated user.

```graphql
MyProfile: Profile!
```

### `Me`

Returns the authenticated user account.

```graphql
Me: User!
```

## Mutations

### `Login`

Authenticates a user with email and password, then returns tokens and user details.

```graphql
Login(input: LoginInput!): AuthResponse!
```

Arguments:

```graphql
input LoginInput {
  email: String!
  password: String!
  meta: RequestMetaInput!
}
```

### `Register`

Creates a new user account and returns tokens, user, and profile details.

```graphql
Register(input: RegisterInput!): AuthResponse!
```

Arguments:

```graphql
input RegisterInput {
  email: String!
  password: String!
  firstName: String!
  lastName: String!
  meta: RequestMetaInput!
}
```

### `Logout`

Logs out the current authenticated session.

```graphql
Logout(input: LogoutInput!): MutationResponse!
```

Arguments:

```graphql
input LogoutInput {
  meta: RequestMetaInput!
}
```

### `RequestPasswordReset`

Starts the password reset flow for a user email.

```graphql
RequestPasswordReset(input: RequestPasswordResetInput!): MutationResponse!
```

Arguments:

```graphql
input RequestPasswordResetInput {
  email: String!
  meta: RequestMetaInput!
}
```

### `ResetPassword`

Completes a password reset using a reset token and a new password.

```graphql
ResetPassword(input: ResetPasswordInput!): MutationResponse!
```

Arguments:

```graphql
input ResetPasswordInput {
  token: String!
  newPassword: String!
}
```

### `RequestEmailVerification`

Starts the email verification flow for the supplied email address.

```graphql
RequestEmailVerification(input: RequestEmailVerificationInput!): MutationResponse!
```

Arguments:

```graphql
input RequestEmailVerificationInput {
  email: String!
  meta: RequestMetaInput!
}
```

### `VerifyEmail`

Verifies a user email using the verification token.

```graphql
VerifyEmail(input: VerifyEmailInput!): MutationResponse!
```

Arguments:

```graphql
input VerifyEmailInput {
  token: String!
  meta: RequestMetaInput!
}
```

### `RevokeSession`

Revokes a specific session using a scoped token and the target token to revoke.

```graphql
RevokeSession(input: RevokeSessionInput): MutationResponse!
```

Arguments:

```graphql
input RevokeSessionInput {
  scopedToken: String!
  tokenToRevoke: String!
}
```

### `RevokeAllSessions`

Revokes all sessions for the authenticated user.

```graphql
RevokeAllSessions(input: ScopedTokenInput!): MutationResponse!
```

Arguments:

```graphql
input ScopedTokenInput {
  scopedToken: String!
  meta: RequestMetaInput!
}
```

### `RevokeOthersSession`

Revokes every session except the current one.

```graphql
RevokeOthersSession(input: RevokeOthersSessionInput!): MutationResponse!
```

Arguments:

```graphql
input RevokeOthersSessionInput {
  scopedToken: String!
  token: String!
}
```

### `GenerateScopedToken`

Generates a short-lived scoped token for sensitive operations such as password changes or session management.

```graphql
GenerateScopedToken(input: GenerateScopedTokenInput!): ScopedTokenResponse!
```

Arguments:

```graphql
input GenerateScopedTokenInput {
  scope: TokenScope!
  authMethod: AuthMethod!
  authValue: String!
  meta: RequestMetaInput!
}
```

### `UpdateProfileAvatarUrl`

Updates the avatar URL for a profile and returns the updated profile.

```graphql
UpdateProfileAvatarUrl(input: UpdateProfileAvatarUrlInput!): Profile!
```

Arguments:

```graphql
input UpdateProfileAvatarUrlInput {
  profileId: ID!
  avatarUrl: String!
}
```

### `UpdateProfileName`

Updates the first and last name for a profile and returns the updated profile.

```graphql
UpdateProfileName(input: UpdateProfileNameInput!): Profile!
```

Arguments:

```graphql
input UpdateProfileNameInput {
  profileId: ID!
  firstName: String!
  lastName: String!
}
```

### `ChangePassword`

Changes the user password using a scoped token and a new password.

```graphql
ChangePassword(input: ChangePasswordInput!): MutationResponse!
```

Arguments:

```graphql
input ChangePasswordInput {
  token: String!
  newPassword: String!
  meta: RequestMetaInput!
}
```

## Response Types

### `User`

Represents the core user account.

```graphql
type User {
  id: ID!
  email: String!
  isEmailVerified: Boolean!
  emailVerifiedAt: DateTime
  twoFactorEnabled: Boolean!
  twoFactorEnabledAt: DateTime
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### `Profile`

Represents the user profile record.

```graphql
type Profile {
  id: ID!
  userId: ID!
  firstName: String!
  lastName: String!
  createdAt: DateTime!
  updatedAt: DateTime!
  avatarUrl: String
}
```

### `Session`

Represents an authenticated session.

```graphql
type Session {
  id: ID!
  userId: ID!
  refreshToken: String!
  deviceId: String!
  ipAddr: String!
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### `Token`

Represents a token value and its expiry time.

```graphql
type Token {
  value: String!
  expiresAt: DateTime!
}
```

### `AuthTokens`

Contains the access and refresh tokens returned after authentication.

```graphql
type AuthTokens {
  accessToken: Token!
  refreshToken: Token!
}
```

### `AuthResponse`

Standard authentication response containing tokens and user data.

```graphql
type AuthResponse {
  tokens: AuthTokens!
  user: User!
  profile: Profile!
}
```

### `MutationResponse`

Generic response for mutations that return operation status.

```graphql
type MutationResponse {
  success: Boolean!
  message: String!
}
```

### `ScopedTokenResponse`

Response type used when the API generates a scoped token.

```graphql
type ScopedTokenResponse {
  token: Token!
}
```

## Enums

### `TokenScope`

Defines the purpose of a scoped token.

```graphql
enum TokenScope {
  AUTH
  REFRESH
  CHANGE_EMAIL
  CHANGE_PASSWORD
  DELETE_ACCOUNT
  RESET_PASSWORD
  VERIFY_EMAIL
  ENABLE_2FA
  DISABLE_2FA
  ADMIN
  IMPERSONATE
  RECOVERY
}
```

### `AuthMethod`

Defines the authentication method used to verify the user before issuing a scoped token.

```graphql
enum AuthMethod {
  PASSWORD
  BIOMETRIC
  OTP
  TOTP
  RECOVERY
  MAGIC_LINK
  SOCIAL_GOOGLE
  SOCIAL_GITHUB
  API_KEY
  SERVICE_ACCOUNT
}
```

/** Client-side account view (derived from `User` for UI) */
export interface Account {
  id: string
  email: string
  createdAt: string
  updatedAt: string
  passwordHash?: string
}

export interface AccountSecurity {
  id: string
  accountId: string
  isEmailVerified: boolean
  emailVerifiedAt: string | null
  twoFactorEnabled: boolean
  twoFactorEnabledAt: string | null
  createdAt: string
  updatedAt: string
}

export interface Oauth {
  id: string
  accountId: string
  provider: string
  subject: string
  token: string
  createdAt: string
  updatedAt: string
}

/** Matches GraphQL `User` */
export interface User {
  id: string
  email: string
  isEmailVerified: boolean
  emailVerifiedAt: string | null
  twoFactorEnabled: boolean
  twoFactorEnabledAt: string | null
  createdAt: string
  updatedAt: string
}

/** Matches GraphQL `Profile` (API uses `userId`) */
export interface Profile {
  id: string
  userId: string
  firstName: string
  lastName: string
  createdAt: string
  updatedAt: string
  avatarUrl: string | null
  /** Convenience alias for existing UI (mirrors `userId`) */
  accountId: string
  displayName: string | null
}

/** Matches GraphQL `Session` */
export interface Session {
  id: string
  userId: string
  deviceId: string
  ipAddr: string
  createdAt: string
  updatedAt: string
}

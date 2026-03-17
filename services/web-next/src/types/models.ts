export interface Account {
  id: string;
  email: string;
  passwordHash: string;
  createdAt: string;
  updatedAt: string;
}

export interface AccountSecurity {
  id: string;
  accountId: string;
  isEmailVerified: boolean;
  emailVerifiedAt: string | null;
  twoFactorEnabled: boolean;
  twoFactorEnabledAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Oauth {
  id: string;
  accountId: string;
  provider: string;
  subject: string;
  token: string;
  createdAt: string;
  updatedAt: string;
}

export interface Profile {
  id: string;
  accountId: string;
  firstName: string;
  lastName: string;
  displayName: string | null;
  avatarUrl: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Session {
  id: string;
  accountId: string;
  refreshToken: string;
  userAgent: string;
  ipAddr: string;
  deviceId: string;
  lastLoginAt: string;
  isRevoked: boolean;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
}

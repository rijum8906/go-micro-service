
export interface Token {
  accessToken: string;
  refreshToken: string;
}
export interface Profile {
  firstName: string;
  lastName: string;
  age: number;
  avatarUrl?: string; // Optional for new accounts
}

export interface Account {
  email: string;
  profiles: Profile[];
  currentProfileIndex: number; // For switching between sub-profiles
}

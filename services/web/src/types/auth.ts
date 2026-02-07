// @/types/auth.ts

export interface Token {
  access_token: string;
  refresh_token: string;
}

export interface Profile {
  id: string; // UUID from Go
  account_id: string; // UUID from Go
  first_name: string;
  last_name: string;
  display_name: string;
  avatar_url: string;
  created_at: string;
  updated_at: string;
}

export interface Account {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

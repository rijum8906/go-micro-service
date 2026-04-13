import type { Token } from '#/types/auth'
import type { Account, Profile } from '#/types/models'
import { create } from 'zustand';

interface AuthState {
  account: Account | null;
  profiles: Profile[] | null;
  token: Token | null;
  activeProfileId: string | null;
  isSignedIn: boolean;
} 

interface AuthActions {
  // Account
  createAccount(data: Account): void;
  updateAccount(data: Partial<Account>): void;
  deleteAccount(): void;

  // Profile
  createProfile(data: Profile): void;
  updateProfile(id: string, data: Partial<Profile>): void;
  deleteProfile(id: string): void;

  // Active profile
  activeProfile(): Profile | null;
  changeActiveProfile(id: string): void;
  switchActiveProfile(): void;

  // Token
  createToken(data: Token): void;
  updateToken(data: Partial<Token>): void;
  deleteToken(): void;

  // Auth
  logout(): void;
}

export const useAuthStore = create<AuthState & AuthActions>((set, get) => ({
  account: null,
  profiles: null,
  activeProfileId: null,
  isSignedIn: false,
  token: null,

  /* ---------- Account ---------- */

  createAccount: (data) => set(() => ({ account: data, isSignedIn: true })),

  updateAccount: (data) =>
    set((state) => ({
      account: state.account ? { ...state.account, ...data } : null,
    })),

  deleteAccount: () =>
    set(() => ({
      account: null,
      isSignedIn: false,
      profiles: null,
      activeProfileId: null,
    })),

  /* ---------- Profiles ---------- */

  createProfile: (data) =>
    set((state) => ({
      profiles: state.profiles ? [data, ...state.profiles] : [data],
      activeProfileId: state.activeProfileId ?? data.id,
    })),

  updateProfile: (id, data) =>
    set((state) => ({
      profiles:
        state.profiles?.map((p) => (p.id === id ? { ...p, ...data } : p)) ??
        null,
    })),

  deleteProfile: (id) =>
    set((state) => {
      const profiles = state.profiles?.filter((p) => p.id !== id) ?? null;
      const activeProfileId =
        state.activeProfileId === id
          ? (profiles?.[0]?.id ?? null)
          : state.activeProfileId;

      return { profiles, activeProfileId };
    }),

  /* ---------- Active profile ---------- */

  activeProfile: () => {
    const { profiles, activeProfileId } = get();
    if (!profiles || !activeProfileId) return null;
    return profiles.find((p) => p.id === activeProfileId) ?? null;
  },

  changeActiveProfile: (id) =>
    set((state) => {
      if (!state.profiles?.some((p) => p.id === id)) return state;
      return { activeProfileId: id };
    }),

  switchActiveProfile: () => {
    const { profiles, activeProfileId } = get();
    if (!profiles || profiles.length < 2) return;

    const currentIndex = profiles.findIndex((p) => p.id === activeProfileId);

    const nextIndex =
      currentIndex === -1 || currentIndex === profiles.length - 1
        ? 0
        : currentIndex + 1;

    set({ activeProfileId: profiles[nextIndex].id });
  },

  /* ---------- Token ---------- */
  createToken: (data) => set(() => ({ token: data })),
  updateToken: (data) =>
    set((state) => ({
      token: state.token ? { ...state.token, ...data } : null,
    })),
  deleteToken: () => set(() => ({ token: null })),

  /* ---------- Auth ---------- */

  logout: () =>
    set(() => ({
      account: null,
      profiles: null,
      activeProfileId: null,
      isSignedIn: false,
    })),
}));

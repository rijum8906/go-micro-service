import type { AuthTokens } from '#/types/auth'
import type { Account, Profile, Session, User } from '#/types/models'
import type { AuthSuccessPayload, SessionBootstrapData } from '#/types/response'
import { create } from 'zustand'

const TOKEN_STORAGE_KEY = 'relay_auth_tokens'

function loadStoredTokens(): AuthTokens | null {
  try {
    const raw = localStorage.getItem(TOKEN_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as AuthTokens
    if (
      parsed?.accessToken?.value &&
      parsed?.refreshToken?.value
    ) {
      return parsed
    }
    return null
  } catch {
    return null
  }
}

function profileFromGql(p: {
  id: string
  userId: string
  firstName: string
  lastName: string
  createdAt: string
  updatedAt: string
  avatarUrl: string | null
}): Profile {
  const display = `${p.firstName} ${p.lastName}`.trim()
  return {
    id: p.id,
    userId: p.userId,
    accountId: p.userId,
    firstName: p.firstName,
    lastName: p.lastName,
    createdAt: p.createdAt,
    updatedAt: p.updatedAt,
    avatarUrl: p.avatarUrl,
    displayName: display || null,
  }
}

function userToAccount(u: User): Account {
  return {
    id: u.id,
    email: u.email,
    createdAt: u.createdAt,
    updatedAt: u.updatedAt,
  }
}

interface AuthState {
  account: Account | null
  profiles: Profile[] | null
  /** Access + refresh tokens (`AuthTokens`) */
  token: AuthTokens | null
  activeProfileId: string | null
  isSignedIn: boolean
  currentSession: Session | null
  /** True after bootstrap finishes (no token, or session validated / cleared) */
  authReady: boolean
}

interface AuthActions {
  createAccount(data: Account): void
  updateAccount(data: Partial<Account>): void
  deleteAccount(): void

  createProfile(data: Profile): void
  updateProfile(id: string, data: Partial<Profile>): void
  deleteProfile(id: string): void

  activeProfile(): Profile | null
  changeActiveProfile(id: string): void
  switchActiveProfile(): void

  createToken(data: AuthTokens): void
  updateToken(data: Partial<AuthTokens>): void
  deleteToken(): void

  applyAuthSuccess(payload: AuthSuccessPayload): void
  applySessionBootstrap(data: SessionBootstrapData): void
  clearAuth(): void
  setAuthReady(ready: boolean): void

  getAccessTokenValue(): string | undefined

  logout(): void
}

export const useAuthStore = create<AuthState & AuthActions>((set, get) => ({
  account: null,
  profiles: null,
  activeProfileId: null,
  isSignedIn: false,
  token: loadStoredTokens(),
  currentSession: null,
  authReady: false,

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
      const profiles = state.profiles?.filter((p) => p.id !== id) ?? null
      const activeProfileId =
        state.activeProfileId === id
          ? (profiles?.[0]?.id ?? null)
          : state.activeProfileId

      return { profiles, activeProfileId }
    }),

  activeProfile: () => {
    const { profiles, activeProfileId } = get()
    if (!profiles || !activeProfileId) return null
    return profiles.find((p) => p.id === activeProfileId) ?? null
  },

  changeActiveProfile: (id) =>
    set((state) => {
      if (!state.profiles?.some((p) => p.id === id)) return state
      return { activeProfileId: id }
    }),

  switchActiveProfile: () => {
    const { profiles, activeProfileId } = get()
    if (!profiles || profiles.length < 2) return

    const currentIndex = profiles.findIndex((p) => p.id === activeProfileId)

    const nextIndex =
      currentIndex === -1 || currentIndex === profiles.length - 1
        ? 0
        : currentIndex + 1

    set({ activeProfileId: profiles[nextIndex].id })
  },

  createToken: (data) => set(() => ({ token: data })),

  updateToken: (data) =>
    set((state) => ({
      token: state.token
        ? {
            ...state.token,
            ...data,
            accessToken: data.accessToken ?? state.token.accessToken,
            refreshToken: data.refreshToken ?? state.token.refreshToken,
          }
        : null,
    })),

  deleteToken: () => set(() => ({ token: null })),

  applyAuthSuccess: (payload) =>
    set(() => {
      const profile = profileFromGql(payload.profile)
      return {
        token: payload.tokens,
        account: userToAccount(payload.user),
        profiles: [profile],
        activeProfileId: profile.id,
        currentSession: null,
        isSignedIn: true,
      }
    }),

  applySessionBootstrap: (data) =>
    set(() => {
      const profile = profileFromGql(data.MyProfile)
      return {
        account: userToAccount(data.Me),
        profiles: [profile],
        activeProfileId: profile.id,
        currentSession: data.GetCurrentSession,
        isSignedIn: true,
      }
    }),

  clearAuth: () =>
    set(() => ({
      account: null,
      profiles: null,
      activeProfileId: null,
      isSignedIn: false,
      token: null,
      currentSession: null,
    })),

  setAuthReady: (ready) => set(() => ({ authReady: ready })),

  getAccessTokenValue: () => get().token?.accessToken?.value,

  logout: () => {
    get().clearAuth()
  },
}))

useAuthStore.subscribe((state, prev) => {
  if (state.token === prev.token) return
  if (state.token) {
    localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(state.token))
  } else {
    localStorage.removeItem(TOKEN_STORAGE_KEY)
  }
})

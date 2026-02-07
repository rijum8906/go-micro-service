import { Account, Profile, Token } from '@/types/auth';
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface AuthState {
  _hasHydrated: boolean;
  account: Account | null;
  profiles: Profile[] | null;
  currentProfileIdx: number | null;
  isSignedIn: boolean;
  token: Token | null;
}

interface AuthActions {
  setHasHydrated: (state: boolean) => void;
  setAuth: (account: Account, profiles: Profile[], token: Token) => void;
  logout: () => void;
  addProfile: (profile: Profile) => void;
  updateProfile: (profileId: string, updates: Partial<Profile>) => void;
  updateAccount: (updates: Partial<Account>) => void;
  setCurrentProfile: (idx: number) => void;
}

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set) => ({
      // Initial State
      _hasHydrated: false,
      account: null,
      profiles: null,
      currentProfileIdx: null,
      isSignedIn: false,
      token: null,

      // Actions
      setHasHydrated: (state) => set({ _hasHydrated: state }),
      setAuth: (account, profiles, token) => set({
        account,
        profiles,
        token,
        isSignedIn: true,
        currentProfileIdx: profiles.length > 0 ? 0 : null,
      }),

      logout: () => {
        set({
          account: null,
          profiles: null,
          token: null,
          isSignedIn: false,
          currentProfileIdx: null
        });
        // Explicitly clear storage for security
        localStorage.removeItem('auth-storage');
      },

      addProfile: (profile) => set((state) => {
        const newProfiles = state.profiles ? [...state.profiles, profile] : [profile];
        return {
          profiles: newProfiles,
          // If this is the first profile added, set it as current
          currentProfileIdx: state.currentProfileIdx === null ? 0 : state.currentProfileIdx
        };
      }),

      updateProfile: (profileId, updates) => set((state) => ({
        // Map through profiles to find matching ID from Go backend
        profiles: state.profiles
          ? state.profiles.map((p) => p.id === profileId ? { ...p, ...updates } : p)
          : null
      })),

      updateAccount: (updates) => set((state) => ({
        // Update account fields like email or metadata
        account: state.account ? { ...state.account, ...updates } : null
      })),

      setCurrentProfile: (idx) => set({ currentProfileIdx: idx }),
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      // Persist essential auth data
      partialize: (state) => ({
        account: state.account,
        profiles: state.profiles,
        token: state.token,
        isSignedIn: state.isSignedIn,
        currentProfileIdx: state.currentProfileIdx
      }),
      onRehydrateStorage: () => (state) => {
        // This runs automatically when the page loads
        state?.setHasHydrated(true);
      },
    }
  )
);

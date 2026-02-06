import { Account, Profile, Token } from '@/types/auth';
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

// 1. Types & Interfaces
interface AuthState {
  account: Account | null;
  isSignedIn: boolean;
  token: Token | null; // Essential for Axios/Go API requests
}

interface AuthActions {
  // Authentication
  setAuth: (account: Account, token: Token) => void;
  logout: () => void;

  // Profile Management
  addProfile: (profile: Profile) => void;
  switchProfile: (index: number) => void;
  updateAccount: (updates: Partial<Account>) => void;
}

// 2. The Store with Persistence
export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set) => ({
      // Initial State
      account: null,
      isSignedIn: false,
      token: null,

      // Set auth data after successful sign-in from Go backend
      setAuth: (account, token) => set({
        account,
        token,
        isSignedIn: true
      }),

      // Clear everything on logout
      logout: () => set({
        account: null,
        token: null,
        isSignedIn: false
      }),

      // Add a sub-profile to the existing account
      addProfile: (profile) => set((state) => ({
        account: state.account
          ? { ...state.account, profiles: [...state.account.profiles, profile] }
          : null
      })),

      // Change which profile is currently active
      switchProfile: (index) => set((state) => ({
        account: state.account
          ? { ...state.account, currentProfileIndex: index }
          : null
      })),

      // General update for email or other settings
      updateAccount: (updates) => set((state) => ({
        account: state.account ? { ...state.account, ...updates } : null
      })),
    }),
    {
      name: 'auth-storage', // Key in localStorage
      storage: createJSONStorage(() => localStorage),
    }
  )
);

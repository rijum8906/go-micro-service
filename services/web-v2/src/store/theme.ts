import { useEffect, useState } from 'react';
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

type Theme = 'light' | 'dark' | 'system';
type ResolvedTheme = 'light' | 'dark';

const SYSTEM_THEME_QUERY = '(prefers-color-scheme: dark)';

interface ThemeStore {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

function getSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined') {
    return 'light';
  }

  return window.matchMedia(SYSTEM_THEME_QUERY).matches ? 'dark' : 'light';
}

export function resolveTheme(theme: Theme): ResolvedTheme {
  return theme === 'system' ? getSystemTheme() : theme;
}

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set) => ({
      theme: 'system', // Default to system for better UX
      setTheme: (theme) => set({ theme }),
      toggleTheme: () =>
        set((state) => ({
          theme: state.theme === 'light' ? 'dark' : 'light',
        })),
    }),
    {
      name: 'theme-storage', // Unique name in localStorage
      storage: createJSONStorage(() => localStorage),
    }
  )
);

export function useResolvedTheme() {
  const theme = useThemeStore((state) => state.theme);
  const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>(() => resolveTheme(theme));

  useEffect(() => {
    if (theme !== 'system') {
      setResolvedTheme(theme);
      return;
    }

    const mediaQueryList = window.matchMedia(SYSTEM_THEME_QUERY);

    const updateResolvedTheme = () => {
      setResolvedTheme(mediaQueryList.matches ? 'dark' : 'light');
    };

    updateResolvedTheme();
    mediaQueryList.addEventListener('change', updateResolvedTheme);

    return () => {
      mediaQueryList.removeEventListener('change', updateResolvedTheme);
    };
  }, [theme]);

  return { theme, resolvedTheme };
}

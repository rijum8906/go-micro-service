import { LoadingScreen } from '@/components/layout/loader'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/store/auth'
import { useThemeStore } from '@/store/theme'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect } from 'react'

export const Route = createFileRoute('/')({
  component: App,
})

function App() {
  const { theme, toggleTheme } = useThemeStore()
  const { _hasHydrated } = useAuthStore();

  useEffect(() => {
    const root = window.document.documentElement
    root.classList.remove('light', 'dark')

    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
      root.classList.add(systemTheme)
    } else {
      root.classList.add(theme)
    }
  }, [theme])

  if (!_hasHydrated) {
    return <LoadingScreen />;
  }
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-2xl font-bold">Current Theme: {theme}</h1>
      <Button onClick={toggleTheme}>Toggle Theme</Button>
      <Link to='/auth/signin'>Signin</Link>
    </div>
  )
}

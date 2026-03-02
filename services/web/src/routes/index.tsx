import { Button } from '@/components/ui/button'
import { useThemeStore } from '@/store/theme'
import { createFileRoute, Link } from '@tanstack/react-router'
import { createServerFn } from '@tanstack/react-start'
import { useEffect } from 'react'

export const Route = createFileRoute('/')({
  component: App,
  loader: async () => {
    const apiUrl = await getApiUrlFn();
    return { apiUrl };
  },
})


const getApiUrlFn = createServerFn({ method: 'GET' }).handler(async () => {
  return process.env.API_BASE_URL || 'not found'
})


function App() {
  const { theme, toggleTheme } = useThemeStore()
  const { apiUrl } = Route.useLoaderData();

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

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-2xl font-bold">Current Theme: {theme}</h1>
      <Button onClick={toggleTheme}>Toggle Theme</Button>
      <Link to='/auth/signin'>Signin</Link>
      <h1>{import.meta.env.VITE_API_BASE_URL || "VITE_API_BASE_URL"}</h1>
      <h1>{import.meta.env.API_BASE_URL || "API_BASE_URL"}</h1>
      <h1>{apiUrl}</h1>
    </div>
  )
}

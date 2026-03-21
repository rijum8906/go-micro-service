import { useEffect } from 'react'
import { Outlet, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { Toaster } from 'sonner'
import { ThemeToggle } from '#/components/ThemeToggle'
import { useResolvedTheme } from '#/store/theme'
import '../styles.css'

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  const { resolvedTheme } = useResolvedTheme()

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedTheme
    document.documentElement.style.colorScheme = resolvedTheme
  }, [resolvedTheme])

  return (
    <>
      <Toaster position="top-center" />
      <div className="fixed right-6 top-6 z-50 sm:right-8 sm:top-8">
        <ThemeToggle />
      </div>
      <Outlet />
      <TanStackDevtools
        config={{ position: 'bottom-right' }}
        plugins={[{ name: 'TanStack Router', render: <TanStackRouterDevtoolsPanel /> }]}
      />
    </>
  )
}
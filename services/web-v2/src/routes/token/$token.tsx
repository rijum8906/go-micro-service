import { createFileRoute } from '@tanstack/react-router'
import { ThemeToggle } from '#/components/ThemeToggle'
import { useThemeStore } from '#/store/theme'
import { ResetPasswordForm } from '#/components/ResetPasswordForm'

export const Route = createFileRoute('/token/$token')({
  component: TokenResetPasswordPage,
})

/** Password reset from emailed link: /token/{scopedToken} */
function TokenResetPasswordPage() {
  const { token: tokenParam } = Route.useParams()
  const { theme } = useThemeStore()
  const isDark = theme === 'dark'

  const scopedToken = (() => {
    try {
      return decodeURIComponent(tokenParam)
    } catch {
      return tokenParam
    }
  })()

  return (
    <div className={`min-h-screen flex flex-col transition-colors duration-300 ${isDark ? 'bg-[#262624]' : 'bg-[#F2EDE4]'}`}>
      <header className="px-12 py-8 flex items-center justify-between">
        <img src="/Logo.svg" alt="Relay" className="h-10" />
        <ThemeToggle />
      </header>

      <div className="flex-1 flex flex-col items-center justify-center px-4 gap-8 -mt-40">
        <h1 className="text-6xl text-[#C97D4E]">
          New password
        </h1>

        <ResetPasswordForm scopedToken={scopedToken} />
      </div>
    </div>
  )
}

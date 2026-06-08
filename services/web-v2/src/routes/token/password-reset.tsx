import { createFileRoute } from '@tanstack/react-router'
import z from 'zod'
import { ThemeToggle } from '#/components/ThemeToggle'
import { useIsDarkTheme } from '#/store/theme'
import { ResetPasswordForm } from '#/components/ResetPasswordForm'

const searchSchema = z.object({
  token: z.string().optional().catch(''),
})

export const Route = createFileRoute('/token/password-reset')({
  component: PasswordResetPage,
  validateSearch: searchSchema,
})

/** Password reset from emailed link (CHANGE_PASSWORD scope): /token/password-reset?token={scopedToken} */
function PasswordResetPage() {
  const { token: tokenSearch } = Route.useSearch()
  const isDark = useIsDarkTheme()

  const scopedToken = (() => {
    const raw = tokenSearch ?? ''
    try {
      return decodeURIComponent(raw)
    } catch {
      return raw
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

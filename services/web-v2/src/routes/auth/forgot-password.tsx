import { createFileRoute, Link } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { toast } from 'sonner'
import { requestPasswordReset } from '#/api/auth'
import { requestPasswordResetSchema, type RequestPasswordResetSchemaType } from '#/schemas/auth'
import { ThemeToggle } from '#/components/ThemeToggle'
import { useIsDarkTheme } from '#/store/theme'

export const Route = createFileRoute('/auth/forgot-password')({
  component: ForgotPasswordPage,
})

function ForgotPasswordPage() {
  const isDark = useIsDarkTheme()

  const form = useForm({
    defaultValues: { email: '' } as RequestPasswordResetSchemaType,
    onSubmit: async ({ value }) => {
      const response = await requestPasswordReset(value)
      if (response.success) {
        toast.success('If an account exists for this email, you will receive reset instructions.')
      } else {
        toast.error(response.message || 'Could not request password reset')
      }
    },
  })

  return (
    <div className={`min-h-screen flex flex-col transition-colors duration-300 ${isDark ? 'bg-[#262624]' : 'bg-[#F2EDE4]'}`}>
      <header className="px-12 py-8 flex items-center justify-between">
        <img src="/Logo.svg" alt="Relay" className="h-10" />
        <ThemeToggle />
      </header>

      <div className="flex-1 flex flex-col items-center justify-center px-4 gap-8 -mt-40">
        <h1 className="text-6xl text-[#C97D4E]">
          Forgot password
        </h1>

        <div className={`w-full max-w-sm rounded-2xl p-8 transition-colors duration-300 ${isDark ? 'bg-[#30302E] shadow-[inset_0_0_0_0.2px_#F2EDE4]' : 'bg-white/40 shadow-[inset_0_0_0_0.3px_#9A9A9A] backdrop-blur-sm'}`}>
          <form
            onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }}
            className="flex flex-col gap-5"
          >
            <form.Field name="email" validators={{ onChange: requestPasswordResetSchema.shape.email }}>
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>Email</label>
                  <input
                    type="email"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="name@example.com"
                    autoComplete="email"
                    className={`h-10 px-4 rounded-lg border text-sm outline-none focus:border-[#C97D4E]/50 transition-colors ${isDark ? 'border-[#F2EDE4]/10 bg-[#262624] text-[#F2EDE4] placeholder:text-[#F2EDE4]/30' : 'border-[#262526]/10 bg-[#F2EDE4]/60 text-[#262526] placeholder:text-[#262526]/30'}`}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-[#B85C5C]">
                      {field.state.meta.errors.map((e: unknown) => (e as { message?: string })?.message ?? String(e)).join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <button
                  type="submit"
                  disabled={!canSubmit || isSubmitting}
                  className="h-10 rounded-lg bg-[#C97D4E] text-[#F2EDE4] font-medium text-sm mt-1 cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed hover:opacity-90 transition-opacity"
                >
                  {isSubmitting ? 'Sending…' : 'Send reset link'}
                </button>
              )}
            </form.Subscribe>
          </form>

          <p className={`text-xs text-center mt-5 ${isDark ? 'text-[#F2EDE4]/40' : 'text-[#262526]/40'}`}>
            <Link to="/auth/signin" style={{ color: '#C97D4E' }} className="hover:opacity-80 font-medium transition-opacity">
              Back to sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}

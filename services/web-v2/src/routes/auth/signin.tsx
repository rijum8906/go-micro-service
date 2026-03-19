import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { toast } from 'sonner'
import { useAuthStore } from '#/store/auth'
import { signin } from '#/api/auth'
import { signinSchema, type SigninSchemaType } from '#/schemas/auth'
import { ThemeToggle } from '#/components/ThemeToggle'
import { useThemeStore } from '#/store/theme'
import z from 'zod'

const searchSchema = z.object({
  redirect: z.string().optional().catch(''),
})

export const Route = createFileRoute('/auth/signin')({
  component: SignInPage,
  validateSearch: searchSchema,
})

function SignInPage() {
  const router = useRouter()
  const { redirect } = Route.useSearch()
  const { createToken, createAccount, createProfile, isSignedIn } = useAuthStore()
  const { theme } = useThemeStore()
  const isDark = theme === 'dark'

  if (isSignedIn) {
    router.navigate({ to: redirect || '/' })
  }

  const form = useForm({
    defaultValues: { email: '', password: '' } as SigninSchemaType,
    onSubmit: async ({ value }) => {
      const response = await signin(value) as any
      if (response.success) {
        const payload = response.data.signin
        createAccount({ id: payload.account.id, email: payload.account.email, createdAt: '', updatedAt: '', passwordHash: '' })
        payload.profiles.forEach((p: any) => createProfile({
          id: p.id,
          firstName: p.firstName,
          lastName: p.lastName,
          displayName: p.displayName,
          avatarUrl: p.avatarUrl,
          accountId: payload.account.id,
          createdAt: '',
          updatedAt: '',
        }))
        createToken({ access_token: payload.tokens.accessToken, refresh_token: payload.tokens.refreshToken })
        toast.success('Logged in successfully')
        router.navigate({ to: redirect || '/' })
      } else {
        toast.error(response.message || 'Authentication failed')
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
        <h1 style={{ fontFamily: 'Junicode, Georgia, serif' }} className="text-6xl text-[#C97D4E]">
          Sign in
        </h1>

        <div className={`w-full max-w-sm rounded-2xl p-8 transition-colors duration-300 ${isDark ? 'bg-[#30302E] shadow-[inset_0_0_0_0.2px_#F2EDE4]' : 'bg-white/40 shadow-[inset_0_0_0_0.3px_#9A9A9A] backdrop-blur-sm'}`}>
          <form
            onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }}
            className="flex flex-col gap-5"
          >
            <form.Field name="email" validators={{ onChange: signinSchema.shape.email }}>
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>Email</label>
                  <input
                    type="email"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="name@example.com"
                    className={`h-10 px-4 rounded-lg border text-sm outline-none focus:border-[#C97D4E]/50 transition-colors ${isDark ? 'border-[#F2EDE4]/10 bg-[#262624] text-[#F2EDE4] placeholder:text-[#F2EDE4]/30' : 'border-[#262526]/10 bg-[#F2EDE4]/60 text-[#262526] placeholder:text-[#262526]/30'}`}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-[#B85C5C]">
                      {field.state.meta.errors.map((e: any) => e?.message ?? e).join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="password" validators={{ onChange: signinSchema.shape.password }}>
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>Password</label>
                  <input
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    className={`h-10 px-4 rounded-lg border text-sm outline-none focus:border-[#C97D4E]/50 transition-colors ${isDark ? 'border-[#F2EDE4]/10 bg-[#262624] text-[#F2EDE4] placeholder:text-[#F2EDE4]/30' : 'border-[#262526]/10 bg-[#F2EDE4]/60 text-[#262526] placeholder:text-[#262526]/30'}`}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-[#B85C5C]">
                      {field.state.meta.errors.map((e: any) => e?.message ?? e).join(', ')}
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
                  {isSubmitting ? 'Signing in...' : 'Sign in'}
                </button>
              )}
            </form.Subscribe>
          </form>

          <p className={`text-xs text-center mt-5 ${isDark ? 'text-[#F2EDE4]/40' : 'text-[#262526]/40'}`}>
            Don't have an account?{' '}
            <Link to="/auth/signup" style={{ color: '#C97D4E' }} className="hover:opacity-80 font-medium transition-opacity">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
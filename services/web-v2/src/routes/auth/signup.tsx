import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { toast } from 'sonner'
import { useAuthStore } from '#/store/auth'
import { signup } from '#/api/auth'
import { signupBaseSchema, type SignupSchemaType } from '#/schemas/auth'
import { ThemeToggle } from '#/components/ThemeToggle'
import { useThemeStore } from '#/store/theme'
import z from 'zod'

const searchSchema = z.object({
  redirect: z.string().optional().catch(''),
})

export const Route = createFileRoute('/auth/signup')({
  component: SignUpPage,
  validateSearch: searchSchema,
})

function SignUpPage() {
  const router = useRouter()
  const { redirect } = Route.useSearch()
  const { createToken, createAccount, createProfile, isSignedIn } = useAuthStore()
  const { theme } = useThemeStore()
  const isDark = theme === 'dark'

  if (isSignedIn) {
    router.navigate({ to: redirect || '/' })
  }

  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: '',
    } as SignupSchemaType,
    onSubmit: async ({ value }) => {
      const response = await signup(value) as any
      if (response.success) {
        const payload = response.data.Register
        const accountId = payload.user.id
        createAccount({ id: accountId, email: payload.user.email, createdAt: '', updatedAt: '', passwordHash: '' })
        const displayName = [payload.profile.firstName, payload.profile.lastName].filter(Boolean).join(' ') || null
        createProfile({
          id: payload.profile.id,
          firstName: payload.profile.firstName,
          lastName: payload.profile.lastName,
          displayName,
          avatarUrl: payload.profile.avatarUrl ?? null,
          accountId,
          createdAt: '',
          updatedAt: '',
        })
        createToken({
          access_token: payload.tokens.accessToken.value,
          refresh_token: payload.tokens.refreshToken.value,
        })
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
        <h1 className="text-6xl text-[#C97D4E]">
          Sign up
        </h1>

        <div className={`w-full max-w-sm rounded-2xl p-8 transition-colors duration-300 ${isDark ? 'bg-[#30302E] shadow-[inset_0_0_0_0.2px_#F2EDE4]' : 'bg-white/40 shadow-[inset_0_0_0_0.3px_#9A9A9A] backdrop-blur-sm'}`}>
          <form
            onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }}
            className="flex flex-col gap-5"
          >
            <div className="grid grid-cols-2 gap-3">
              <form.Field name="firstName" validators={{ onChange: signupBaseSchema.shape.firstName }}>
                {(field) => (
                  <div className="flex flex-col gap-1.5">
                    <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>First name</label>
                    <input
                      type="text"
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="John"
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

              <form.Field name="lastName" validators={{ onChange: signupBaseSchema.shape.lastName }}>
                {(field) => (
                  <div className="flex flex-col gap-1.5">
                    <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>Last name</label>
                    <input
                      type="text"
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="Doe"
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
            </div>

            <form.Field name="email" validators={{ onChange: signupBaseSchema.shape.email }}>
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

            <form.Field name="password" validators={{ onChange: signupBaseSchema.shape.password }}>
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>Password</label>
                  <input
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    className={`h-10 px-4 rounded-lg border text-sm outline-none focus:border-[#C97D4E]/50 transition-colors ${isDark ? 'border-[#F2EDE4]/10 bg-[#262624] text-[#F2EDE4] placeholder:text-[#F2EDE4]/30' : 'border-[#262526]/10 bg-[#F2EDE4]/60 text-[#262526]'}`}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-[#B85C5C]">
                      {field.state.meta.errors.map((e: any) => e?.message ?? e).join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field
              name="confirmPassword"
              validators={{
                onChangeListenTo: ['password'],
                onChange: ({ value, fieldApi }) => {
                  if (value !== fieldApi.form.getFieldValue('password')) {
                    return 'Passwords do not match'
                  }
                },
              }}
            >
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>Confirm password</label>
                  <input
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    className={`h-10 px-4 rounded-lg border text-sm outline-none focus:border-[#C97D4E]/50 transition-colors ${isDark ? 'border-[#F2EDE4]/10 bg-[#262624] text-[#F2EDE4] placeholder:text-[#F2EDE4]/30' : 'border-[#262526]/10 bg-[#F2EDE4]/60 text-[#262526]'}`}
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
                  {isSubmitting ? 'Creating account...' : 'Create account'}
                </button>
              )}
            </form.Subscribe>
          </form>

          <p className={`text-xs text-center mt-5 ${isDark ? 'text-[#F2EDE4]/40' : 'text-[#262526]/40'}`}>
            Already have an account?{' '}
            <Link to="/auth/signin" style={{ color: '#C97D4E' }} className="hover:opacity-80 font-medium transition-opacity">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}


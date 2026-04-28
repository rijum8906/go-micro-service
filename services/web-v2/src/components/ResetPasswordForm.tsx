import { Link, useRouter } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { toast } from 'sonner'
import { resetPassword } from '#/api/auth'
import { resetPasswordSchema } from '#/schemas/auth'
import { useIsDarkTheme } from '#/store/theme'

type Props = {
  /** Scoped token from URL path `/token/{scopedToken}` */
  scopedToken: string
}

export function ResetPasswordForm({ scopedToken }: Props) {
  const router = useRouter()
  const isDark = useIsDarkTheme()

  const form = useForm({
    defaultValues: {
      token: scopedToken,
      newPassword: '',
      confirmPassword: '',
    },
    onSubmit: async ({ value }) => {
      const t = value.token?.trim()
      if (!t) {
        toast.error('Reset token is missing. Open the link from your email.')
        return
      }
      const response = await resetPassword({ token: t, newPassword: value.newPassword })
      if (response.success) {
        toast.success('Your password has been updated. You can sign in.')
        router.navigate({ to: '/auth/signin' })
      } else {
        toast.error(response.message || 'Could not reset password')
      }
    },
  })

  return (
    <div className={`w-full max-w-sm rounded-2xl p-8 transition-colors duration-300 ${isDark ? 'bg-[#30302E] shadow-[inset_0_0_0_0.2px_#F2EDE4]' : 'bg-white/40 shadow-[inset_0_0_0_0.3px_#9A9A9A] backdrop-blur-sm'}`}>
      <form
        onSubmit={async (e) => { e.preventDefault(); await form.handleSubmit() }}
        className="flex flex-col gap-5"
      >
        <p className={`text-xs ${isDark ? 'text-[#F2EDE4]/60' : 'text-[#262526]/60'}`}>
          Reset link loaded. Choose your new password below.
        </p>

        <form.Field name="newPassword" validators={{ onChange: resetPasswordSchema.shape.newPassword }}>
          {(field) => (
            <div className="flex flex-col gap-1.5">
              <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>New password</label>
              <input
                type="password"
                value={field.state.value}
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                autoComplete="new-password"
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

        <form.Field
          name="confirmPassword"
          validators={{
            onChangeListenTo: ['newPassword'],
            onChange: ({ value, fieldApi }) => {
              if (value !== fieldApi.form.getFieldValue('newPassword')) {
                return "Passwords don't match"
              }
              return undefined
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
                autoComplete="new-password"
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
              {isSubmitting ? 'Updating…' : 'Update password'}
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
  )
}

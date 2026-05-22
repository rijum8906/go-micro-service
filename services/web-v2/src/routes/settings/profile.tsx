import { createFileRoute, useRouter } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { useAuthStore } from '#/store/auth'
import { useIsDarkTheme } from '#/store/theme'
import { changePasswordMutation, generateScopedToken, updateProfileName } from '#/api/user'
import { updateProfileNameSchema, type UpdateProfileNameSchemaType, changePasswordSchema, type ChangePasswordSchemaType } from '#/schemas/auth'
import { ThemeToggle } from '#/components/ThemeToggle'
import { Avatar } from '#/components/Avatar'
import { Card } from '#/components/Card'
import { FormField } from '#/components/FormField'

export const Route = createFileRoute('/settings/profile')({
  component: ProfilePage,
})

function ProfilePage() {
  const router = useRouter()
  const isDark = useIsDarkTheme()
  const {
    authReady,
    isSignedIn,
    account,
    profiles,
    activeProfileId,
    updateProfile,
    getAccessTokenValue,
  } = useAuthStore()

  const profile = profiles?.find((p) => p.id === activeProfileId) ?? null

  const [passwordSubmitting, setPasswordSubmitting] = useState(false)

  useEffect(() => {
    if (authReady && !isSignedIn) {
      router.navigate({ to: '/auth/signin', search: { redirect: '/settings/profile' } })
    }
  }, [authReady, isSignedIn, router])

  if (!authReady) return null

  if (!isSignedIn || !account || !profile) return null

  const memberSince = new Date(account.createdAt).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  const nameForm = useForm({
    defaultValues: {
      firstName: profile.firstName,
      lastName: profile.lastName,
    } as UpdateProfileNameSchemaType,
    onSubmit: async ({ value }) => {
      const getToken = () => getAccessTokenValue()
      const result = await updateProfileName(
        { profileId: profile.id, firstName: value.firstName, lastName: value.lastName },
        getToken,
      )
      if (result.success && result.data) {
        const { id, userId, firstName, lastName, createdAt, updatedAt, avatarUrl } = result.data.UpdateProfileName
        const display = `${firstName} ${lastName}`.trim()
        updateProfile(profile.id, {
          id,
          userId,
          accountId: userId,
          firstName,
          lastName,
          createdAt,
          updatedAt,
          avatarUrl,
          displayName: display || null,
        })
        toast.success('Profile updated')
      } else {
        toast.error(result.message || 'Failed to update profile')
      }
    },
  })

  const passwordForm = useForm({
    defaultValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    } as ChangePasswordSchemaType,
    validators: {
      onSubmit: ({ value }) => {
        const parsed = changePasswordSchema.safeParse(value)
        if (!parsed.success) {
          return parsed.error.issues[0]?.message ?? 'Invalid password fields'
        }
      },
    },
    onSubmit: async ({ value }) => {
      const parsed = changePasswordSchema.safeParse(value)
      if (!parsed.success) {
        toast.error(parsed.error.issues[0]?.message ?? 'Invalid password fields')
        return
      }
      setPasswordSubmitting(true)
      const getToken = () => getAccessTokenValue()
      const scoped = await generateScopedToken(
        { scope: 'CHANGE_PASSWORD', authMethod: 'PASSWORD', authValue: parsed.data.currentPassword },
        getToken,
      )
      if (!scoped.success || !scoped.data) {
        setPasswordSubmitting(false)
        toast.error(scoped.message || 'Current password is incorrect')
        return
      }
      const scopedTokenValue = scoped.data.GenerateScopedToken.token.value
      const result = await changePasswordMutation(
        { token: scopedTokenValue, newPassword: parsed.data.newPassword },
        getToken,
      )
      setPasswordSubmitting(false)
      if (result.success && result.data?.ChangePassword.success) {
        passwordForm.reset()
        toast.success(result.data.ChangePassword.message || 'Password changed')
      } else {
        const message = result.success
          ? result.data?.ChangePassword.message
          : result.message
        toast.error(message || 'Failed to change password')
      }
    },
  })

  return (
    <div className={`min-h-screen flex flex-col transition-colors duration-300 ${isDark ? 'bg-[#262624]' : 'bg-[#F2EDE4]'}`}>
      <header className="px-12 py-8 flex items-center justify-between">
        <button
          onClick={() => router.navigate({ to: '/dashboard' })}
          className="cursor-pointer"
          aria-label="Back to dashboard"
        >
          <img src="/Logo.svg" alt="Relay" className="h-10" />
        </button>
        <ThemeToggle />
      </header>

      <div className="flex-1 flex flex-col items-center justify-center px-4 gap-8 -mt-20">
        <h1 className="text-4xl text-[#C97D4E]">
          Profile
        </h1>

        <Card isDark={isDark} className="max-w-lg">
          <div className="flex flex-col items-center gap-4 mb-6">
            <Avatar firstName={profile.firstName} lastName={profile.lastName} size="lg" />
            <div className="text-center">
              <p className={`text-lg font-medium ${isDark ? 'text-[#F2EDE4]' : 'text-[#262526]'}`}>
                {profile.firstName} {profile.lastName}
              </p>
              <p className={`text-sm ${isDark ? 'text-[#F2EDE4]/60' : 'text-[#262526]/60'}`}>
                {account.email}
              </p>
              <p className={`text-xs mt-1 ${isDark ? 'text-[#F2EDE4]/40' : 'text-[#262526]/40'}`}>
                Member since {memberSince}
              </p>
            </div>
          </div>
        </Card>

        <Card isDark={isDark} className="max-w-lg">
          <h2 className={`text-lg font-medium mb-5 ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>
            Edit name
          </h2>
          <form
            onSubmit={async (e) => { e.preventDefault(); await nameForm.handleSubmit() }}
            className="flex flex-col gap-5"
          >
            <nameForm.Field name="firstName" validators={{ onChange: updateProfileNameSchema.shape.firstName }}>
              {(field) => (
                <FormField
                  label="First name"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  errors={field.state.meta.errors}
                  isDark={isDark}
                />
              )}
            </nameForm.Field>

            <nameForm.Field name="lastName" validators={{ onChange: updateProfileNameSchema.shape.lastName }}>
              {(field) => (
                <FormField
                  label="Last name"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  errors={field.state.meta.errors}
                  isDark={isDark}
                />
              )}
            </nameForm.Field>

            <nameForm.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <button
                  type="submit"
                  disabled={!canSubmit || isSubmitting}
                  className="h-10 rounded-lg bg-[#C97D4E] text-[#F2EDE4] font-medium text-sm mt-1 cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed hover:opacity-90 transition-opacity"
                >
                  {isSubmitting ? 'Saving...' : 'Save changes'}
                </button>
              )}
            </nameForm.Subscribe>
          </form>
        </Card>

        <Card isDark={isDark} className="max-w-lg">
          <h2 className={`text-lg font-medium mb-5 ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>
            Change password
          </h2>
          <form
            onSubmit={async (e) => { e.preventDefault(); await passwordForm.handleSubmit() }}
            className="flex flex-col gap-5"
          >
            <passwordForm.Field name="currentPassword" validators={{ onChange: changePasswordSchema.shape.currentPassword }}>
              {(field) => (
                <FormField
                  label="Current password"
                  type="password"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  errors={field.state.meta.errors}
                  isDark={isDark}
                />
              )}
            </passwordForm.Field>

            <passwordForm.Field name="newPassword" validators={{ onChange: changePasswordSchema.shape.newPassword }}>
              {(field) => (
                <FormField
                  label="New password"
                  type="password"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  errors={field.state.meta.errors}
                  isDark={isDark}
                />
              )}
            </passwordForm.Field>

            <passwordForm.Field
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
                <FormField
                  label="Confirm password"
                  type="password"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  errors={field.state.meta.errors}
                  isDark={isDark}
                />
              )}
            </passwordForm.Field>

            <passwordForm.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <button
                  type="submit"
                  disabled={!canSubmit || isSubmitting || passwordSubmitting}
                  className="h-10 rounded-lg bg-[#C97D4E] text-[#F2EDE4] font-medium text-sm mt-1 cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed hover:opacity-90 transition-opacity"
                >
                  {isSubmitting || passwordSubmitting ? 'Changing password...' : 'Change password'}
                </button>
              )}
            </passwordForm.Subscribe>

            {passwordForm.state.fieldMeta.currentPassword?.errors.length === 0 &&
              passwordForm.state.fieldMeta.newPassword?.errors.length === 0 &&
              passwordForm.state.fieldMeta.confirmPassword?.errors.length === 0 &&
              passwordForm.state.errors.length > 0 && (
              <p className="text-xs text-[#B85C5C]">
                {passwordForm.state.errors.join(', ')}
              </p>
            )}
          </form>
        </Card>
      </div>
    </div>
  )
}
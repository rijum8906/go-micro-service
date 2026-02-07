import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { SignupRequest, signupSchema } from '@/schemas/auth'
import { SocialAuth } from '@/components/auth/social-auth'
import { toast } from 'sonner'
import { AxiosError } from 'axios'
import { generateDeviceId } from '@/lib/device'
import { api } from '@/api/axios'
import { AuthResponse } from '@/types/response'
import { useAuthStore } from '@/store/auth'

export const Route = createFileRoute('/auth/signup')({
  component: SignUpComponent,
})

function SignUpComponent() {
  const router = useRouter()
  const { setAuth } = useAuthStore()

  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
    validators: { onChange: signupSchema },
    onSubmit: async ({ value }) => {
      try {
        const payload: SignupRequest = {
          ...value,
          metadata: { deviceId: generateDeviceId() }
        }
        const response = await api.post<AuthResponse>('/auth/signup', payload, {
          timeout: 5000,
        })
        if (!response.data.data.account || !response.data.data.token) {
          toast.error(response.data.message || 'Registration failed. Please try again.')
          router.navigate({ to: '/auth/signin' })
          setAuth(response.data.data.account, response.data.data.profiles, response.data.data.token)
          return
        } else {
          toast.success(response.data.message || 'Registration successful')
          router.navigate({ to: '/auth/signin' })
        }

      } catch (err) {
        const error = err as AxiosError<{ message?: string }>

        if (!error.response) {
          toast.error('Network error: Registration service is unreachable')
          return
        }

        const serverMessage = error.response.data?.message
        toast.error(serverMessage || 'Registration failed. Please try again.')
      }
    },
  })

  const ErrorDisplay = ({ errors }: { errors: any[] }) => {
    if (!errors.length) return null
    return (
      <p className="text-xs text-destructive mt-1">
        {errors.map((err) => (typeof err === 'object' ? err.message : err)).join(', ')}
      </p>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4 py-12">
      <Card className="w-full max-w-lg shadow-xl border-muted/50">
        <CardHeader className="space-y-1 text-center">
          <CardTitle className="text-3xl font-bold tracking-tight">Create an account</CardTitle>
          <CardDescription>Enter your details to join our platform</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
            className="space-y-4"
          >
            <div className="grid grid-cols-2 gap-4">
              <form.Field name="firstName">
                {(field) => (
                  <div className="space-y-1">
                    <Label htmlFor={field.name}>First Name</Label>
                    <Input
                      id={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="John"
                      autoComplete="given-name"
                    />
                    <ErrorDisplay errors={field.state.meta.errors} />
                  </div>
                )}
              </form.Field>
              <form.Field name="lastName">
                {(field) => (
                  <div className="space-y-1">
                    <Label htmlFor={field.name}>Last Name</Label>
                    <Input
                      id={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      placeholder="Doe"
                      autoComplete="family-name"
                    />
                    <ErrorDisplay errors={field.state.meta.errors} />
                  </div>
                )}
              </form.Field>
            </div>

            <form.Field name="email">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor={field.name}>Email Address</Label>
                  <Input
                    id={field.name}
                    type="email"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="john@example.com"
                    autoComplete="email"
                  />
                  <ErrorDisplay errors={field.state.meta.errors} />
                </div>
              )}
            </form.Field>

            <form.Field name="password">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor={field.name}>Password</Label>
                  <Input
                    id={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    autoComplete="new-password"
                  />
                  <ErrorDisplay errors={field.state.meta.errors} />
                </div>
              )}
            </form.Field>

            <form.Field name="confirmPassword">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor={field.name}>Confirm Password</Label>
                  <Input
                    id={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    autoComplete="new-password"
                  />
                  <ErrorDisplay errors={field.state.meta.errors} />
                </div>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <Button className="w-full h-11 text-base font-semibold" type="submit" disabled={!canSubmit || isSubmitting}>
                  {isSubmitting ? 'Registering...' : 'Create Account'}
                </Button>
              )}
            </form.Subscribe>
          </form>

          <div className="mt-6">
            <SocialAuth />
          </div>
        </CardContent>
        <CardFooter className="flex justify-center border-t p-6">
          <p className="text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link to="/auth/signin" className="text-primary font-semibold hover:underline">
              Sign in
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}

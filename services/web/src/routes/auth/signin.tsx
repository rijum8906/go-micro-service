import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { loginSchema, LoginSchemaType } from '@/schemas/auth'
import { SocialAuth } from '@/components/auth/social-auth'
import { toast } from 'sonner'
import axios, { AxiosError } from 'axios'
import { generateDeviceId } from '@/lib/device'

export const Route = createFileRoute('/auth/signin')({
  component: SignInComponent,
})

function SignInComponent() {
  const router = useRouter()

  const form = useForm({
    defaultValues: { email: '', password: '' } as LoginSchemaType,
    onSubmit: async ({ value }) => {
      try {
        const payload = {
          ...value,
          metadata: { deviceId: generateDeviceId() }
        }

        const response = await axios.post('http://localhost:8906/api/v1/auth/signin', payload, {
          timeout: 5000,
        })

        toast.success(response.data.message || 'Logged in successfully')

        if (response.data.token) {
          localStorage.setItem('token', response.data.token)
        }

        router.navigate({ to: '/' })
      } catch (err) {
        const error = err as AxiosError<{ message?: string }>

        // Handle network/server-down errors
        if (!error.response) {
          toast.error('Unable to connect to the authentication service')
          return
        }

        // Professional error prioritization using the 'message' field
        const serverMessage = error.response.data?.message
        toast.error(serverMessage || 'An unexpected error occurred during sign in')
      }
    }
  })

  const renderErrors = (field: any) => {
    if (!field.state.meta.errors.length) return null
    return (
      <p className="text-xs text-destructive mt-1">
        {field.state.meta.errors.map((err: any) =>
          typeof err === 'object' ? err.message : err
        ).join(', ')}
      </p>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md shadow-xl border-muted/50">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold tracking-tight">Sign in</CardTitle>
          <CardDescription>Enter your credentials to access your account</CardDescription>
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
            <form.Field name="email" validators={{ onChange: loginSchema.shape.email }}>
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor={field.name}>Email</Label>
                  <Input
                    id={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="name@example.com"
                    autoComplete="email"
                  />
                  {renderErrors(field)}
                </div>
              )}
            </form.Field>

            <form.Field name="password" validators={{ onChange: loginSchema.shape.password }}>
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor={field.name}>Password</Label>
                  <Input
                    id={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    autoComplete="current-password"
                  />
                  {renderErrors(field)}
                </div>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <Button className="w-full h-10" type="submit" disabled={!canSubmit || isSubmitting}>
                  {isSubmitting ? 'Authenticating...' : 'Sign In'}
                </Button>
              )}
            </form.Subscribe>
          </form>

          <div className="mt-6">
            <SocialAuth />
          </div>
        </CardContent>
        <CardFooter className="flex justify-center border-t py-4">
          <p className="text-sm text-muted-foreground">
            Don't have an account? <Link to="/auth/signup" className="text-primary font-medium hover:underline">Sign up</Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}

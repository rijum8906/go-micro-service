import { createFileRoute, Link } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { loginSchema } from '@/schemas/auth'
import { SocialAuth } from '@/components/auth/social-auth'

export const Route = createFileRoute('/auth/signin')({
  component: SignInComponent,
})

function SignInComponent() {
  const form = useForm({
    defaultValues: { email: '', password: '' },
    // Use the zodValidator at the top level
    onSubmit: async ({ value }) => {
      console.log('Form submitted:', value)
    },
  })

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md shadow-lg border-muted">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold tracking-tight">Sign in</CardTitle>
          <CardDescription>Enter your email to access your account</CardDescription>
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
            {/* Field for Email */}
            <form.Field
              name="email"
              validators={{
                onChange: loginSchema.shape.email,
              }}
            >
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor={field.name}>Email</Label>
                  <Input
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="john@gmail.com"
                    autoComplete="email"
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-destructive">
                      {field.state.meta.errors.map((err) =>
                        typeof err === 'object' && err !== null ? (err as any).message : err
                      ).join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            {/* Field for Password */}
            <form.Field
              name="password"
              validators={{
                onChange: loginSchema.shape.password,
              }}
            >
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor={field.name}>Password</Label>
                  <Input
                    id={field.name}
                    name={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    autoComplete='current-password'
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs text-destructive">
                      {field.state.meta.errors.map((err) =>
                        typeof err === 'object' && err !== null ? (err as any).message : err
                      ).join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <Button className="w-full" type="submit" disabled={!canSubmit}>
                  {isSubmitting ? 'Signing in...' : 'Sign In'}
                </Button>
              )}
            </form.Subscribe>
          </form>

          {/* Social Auth */}
          <div className="mt-6">
            <SocialAuth />
          </div>
        </CardContent>
        <CardFooter>
          <p className="text-sm text-muted-foreground">
            Don't have an account?{' '}
            <Link to="/auth/signup" className="text-primary hover:underline font-medium">
              Sign up
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}

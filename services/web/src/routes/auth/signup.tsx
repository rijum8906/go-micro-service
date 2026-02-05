import { createFileRoute, Link } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { signupSchema } from '@/schemas/auth'
import { SocialAuth } from '@/components/auth/social-auth'

export const Route = createFileRoute('/auth/signup')({
  component: SignUpComponent,
})

function SignUpComponent() {
  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: '',
    },
    validators: {
      onChange: signupSchema,
    },
    onSubmit: async ({ value }) => {
      console.log('Registering user:', value)
    },
  })

  // Helper to render error messages correctly
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
      <Card className="w-full max-w-lg shadow-xl border-muted">
        <CardHeader className="space-y-1 text-center">
          <CardTitle className="text-3xl font-bold tracking-tight">Create an account</CardTitle>
          <CardDescription>Enter your details to get started with our platform</CardDescription>
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
            {/* First and Last Name Grid */}
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
                      autoComplete='given-name'
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
                      autoComplete='family-name'
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
                  />
                  <ErrorDisplay errors={field.state.meta.errors} />
                </div>
              )}
            </form.Field>

            <form.Subscribe selector={(state) => [state.canSubmit, state.isSubmitting]}>
              {([canSubmit, isSubmitting]) => (
                <Button className="w-full h-11 text-base font-semibold" type="submit" disabled={!canSubmit}>
                  {isSubmitting ? 'Creating account...' : 'Create Account'}
                </Button>
              )}
            </form.Subscribe>
          </form>


          {/* Social Auth */}
          <div className="mt-6">
            <SocialAuth />
          </div>
        </CardContent>
        <CardFooter className="flex justify-center border-t p-6">
          <p className="text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link to="/auth/signin" className="text-primary hover:underline font-semibold">
              Sign in
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}

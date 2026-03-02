import { createFileRoute, Link, useRouter } from '@tanstack/react-router';
import { useForm } from '@tanstack/react-form';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { signinSchema, type SigninSchemaType } from '@/schemas/auth';
import { SocialAuth } from '@/components/auth/social-auth';
import { toast } from 'sonner';
import z from 'zod';
import { useEffect } from 'react';
import { useAuthStore } from '@/store/auth';
import { signin } from '@/api/auth';
import { getBaseApiUrlFn } from '@/lib/server-api';

const signinSearchSchema = z.object({
  redirect: z.string().optional().catch(''),
});
export const Route = createFileRoute('/auth/signin')({
  component: SignInComponent,
  validateSearch: signinSearchSchema,
  loader: async () => {
    const apiUrl = await getBaseApiUrlFn();
    return { apiUrl };
  },
});

function SignInComponent() {
  const router = useRouter();
  const { redirect } = Route.useSearch();
  const { createToken, createAccount, createProfile, isSignedIn } =
    useAuthStore();
  const { apiUrl } = Route.useLoaderData();

  useEffect(() => {
    if (isSignedIn) {
      router.navigate({ to: redirect || '/' });
    }
  }, [isSignedIn, router.navigate, redirect]);

  const form = useForm({
    defaultValues: { email: '', password: '' } as SigninSchemaType,
    onSubmit: async ({ value }) => {
      const response = await signin(value);
      if (response.success) {
        createAccount(response.data.account);
        response.data.profiles.forEach(createProfile);
        createToken(response.data.token);
        toast.success(response.message || 'Logged in successfully');
        router.navigate({ to: redirect || '/' });
      } else {
        toast.error(response.message || 'Authentication failed');
      }
    },
  });

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md shadow-xl border-muted/50">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold tracking-tight">
            Sign in
          </CardTitle>
          <CardDescription>
            Enter your credentials to access your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              form.handleSubmit();
            }}
            className="space-y-4"
          >
            {/* Email Field */}
            <form.Field
              name="email"
              validators={{ onChange: signinSchema.shape.email }}
            >
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor={field.name}>Email</Label>
                  <Input
                    id={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="name@example.com"
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs font-medium text-destructive">
                      {field.state.meta.errors.join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            {/* Password Field */}
            <form.Field
              name="password"
              validators={{ onChange: signinSchema.shape.password }}
            >
              {(field) => (
                <div className="space-y-2">
                  <Label htmlFor={field.name}>Password</Label>
                  <Input
                    id={field.name}
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p className="text-xs font-medium text-destructive">
                      {field.state.meta.errors.join(', ')}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
            >
              {([canSubmit, isSubmitting]) => (
                <Button
                  className="w-full h-10"
                  type="submit"
                  disabled={!canSubmit || isSubmitting}
                >
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
            Don't have an account?{' '}
            <Link
              to="/auth/signup"
              className="text-primary font-medium hover:underline"
            >
              Sign up
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}

import { useState } from 'react';
import { Eye, EyeOff, Lock, ShieldCheck } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import { createFileRoute } from '@tanstack/react-router';
import { useForm } from '@tanstack/react-form';
import {
  changePasswordSchema,
  type ChangePasswordSchemaType,
} from '@/schemas/auth';

export const Route = createFileRoute('/my-account/edit/_layout/security')({
  component: RouteComponent,
});

function RouteComponent() {
  const [show, setShow] = useState(false);

  const form = useForm({
    defaultValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    } as ChangePasswordSchemaType,
    onSubmit: async ({ value }) => {
      // call API here
      console.log(value);
    },
  });

  return (
    <div className="space-y-10">
      {/* Page header */}
      <div className="space-y-1">
        <h3 className="text-2xl font-semibold">Security</h3>
        <p className="text-sm text-muted-foreground">
          Manage how your account is protected.
        </p>
      </div>

      <Separator />

      {/* Change password section */}
      <section className="rounded-xl border bg-background p-6 space-y-6 max-w-md">
        <div className="flex items-start gap-3">
          <div className="rounded-lg border p-2">
            <ShieldCheck className="h-5 w-5 text-primary" />
          </div>

          <div className="space-y-1">
            <h4 className="text-lg font-medium">Change password</h4>
            <p className="text-sm text-muted-foreground">
              Use a strong password you don’t reuse anywhere else.
            </p>
          </div>
        </div>

        <Separator />

        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          className="space-y-4"
        >
          {/* Current password */}
          <form.Field
            name="currentPassword"
            validators={{
              onChange: changePasswordSchema.shape.currentPassword,
            }}
          >
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor={field.name}>Current password</Label>
                <div className="relative">
                  <Input
                    id={field.name}
                    type={show ? 'text' : 'password'}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="••••••••"
                    autoComplete="current-password"
                  />
                  <PasswordToggle show={show} setShow={setShow} />
                </div>
                {field.state.meta.errors.length > 0 && (
                  <p className="text-xs font-medium text-destructive">
                    {field.state.meta.errors.join(', ')}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          {/* New password */}
          <form.Field
            name="newPassword"
            validators={{
              onChange: changePasswordSchema.shape.newPassword,
            }}
          >
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor={field.name}>New password</Label>
                <div className="relative">
                  <Input
                    id={field.name}
                    type={show ? 'text' : 'password'}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder="At least 8 characters"
                    autoComplete="new-password"
                  />
                  <PasswordToggle show={show} setShow={setShow} />
                </div>
                {field.state.meta.errors.length > 0 && (
                  <p className="text-xs font-medium text-destructive">
                    {field.state.meta.errors[0].message}
                    {console.log(field.state.meta.errors)}
                  </p>
                )}
              </div>
            )}
          </form.Field>

          {/* Confirm password */}
          <form.Field
            name="confirmPassword"
            validators={{
              onChange: changePasswordSchema.shape.confirmPassword,
            }}
          >
            {(field) => (
              <div className="space-y-1">
                <Label htmlFor={field.name}>Confirm new password</Label>
                <Input
                  id={field.name}
                  type={show ? 'text' : 'password'}
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Repeat new password"
                  autoComplete="new-password"
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
                type="submit"
                className="w-full h-10"
                disabled={!canSubmit || isSubmitting}
              >
                {isSubmitting ? (
                  'Updating password…'
                ) : (
                  <>
                    <Lock className="mr-2 h-4 w-4" />
                    Update password
                  </>
                )}
              </Button>
            )}
          </form.Subscribe>
        </form>
      </section>
    </div>
  );
}

function PasswordToggle({
  show,
  setShow,
}: {
  show: boolean;
  setShow: (v: boolean) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => setShow(!show)}
      className={cn(
        'absolute right-3 top-1/2 -translate-y-1/2',
        'text-muted-foreground hover:text-foreground transition',
      )}
    >
      {show ? <EyeOff size={18} /> : <Eye size={18} />}
    </button>
  );
}

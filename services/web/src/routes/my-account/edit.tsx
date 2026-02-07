import { createFileRoute, useRouter } from '@tanstack/react-router';
import { useForm } from '@tanstack/react-form';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { toast } from 'sonner';
import { useAuthStore } from '@/store/auth';
import { useId, useState } from 'react';

export const Route = createFileRoute('/my-account/edit')({
  component: EditAccountComponent,
});

function EditAccountComponent() {
  const id = useId();
  const router = useRouter();
  const { profiles, currentProfileIdx } = useAuthStore();
  const profile = profiles?.[currentProfileIdx ?? 0];

  // Local state for image preview
  const [previewUrl, setPreviewUrl] = useState<string | null>(
    profile?.avatar_url || null,
  );

  const form = useForm({
    defaultValues: {
      firstName: profile?.first_name || '',
      lastName: profile?.last_name || '',
      avatar: null as File | null,
    },
    onSubmit: async ({ value }) => {
      try {
        // Since we have an image, we use FormData instead of a standard JSON payload
        const formData = new FormData();
        formData.append('first_name', value.firstName);
        formData.append('last_name', value.lastName);
        if (value.avatar) {
          formData.append('avatar', value.avatar);
        }

        console.log('Sending update...', value);
        // await api.put('/auth/profile', formData);

        toast.success('Profile updated successfully');
        router.navigate({ to: '/my-account' });
      } catch (err) {
        toast.error('Failed to update profile');
      }
    },
  });

  return (
    <div className="container mx-auto py-10 px-4 max-w-2xl">
      <h1 className="text-3xl font-bold mb-6">Edit Profile</h1>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
        className="space-y-8 p-6 border rounded-lg bg-card shadow-sm"
      >
        {/* Avatar Section */}
        <div className="flex flex-col items-center gap-4 border-b pb-6">
          <Avatar className="h-24 w-24 border-2 border-muted">
            <AvatarImage src={previewUrl || ''} />
            <AvatarFallback>
              {profile?.first_name?.charAt(0)}
              {profile?.last_name?.charAt(0)}
            </AvatarFallback>
          </Avatar>

          <form.Field name="avatar">
            {(field) => (
              <div className="flex flex-col items-center gap-2">
                <Label
                  htmlFor={id}
                  className="cursor-pointer bg-secondary px-4 py-2 rounded-md hover:bg-secondary/80 transition-colors"
                >
                  Change Avatar
                </Label>
                <Input
                  id={id}
                  type="file"
                  className="hidden"
                  accept="image/*"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) {
                      field.handleChange(file);
                      setPreviewUrl(URL.createObjectURL(file));
                    }
                  }}
                />
                {field.state.meta.errors.length > 0 && (
                  <p className="text-xs text-destructive">
                    {field.state.meta.errors.join(', ')}
                  </p>
                )}
              </div>
            )}
          </form.Field>
        </div>

        {/* Text Fields */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <form.Field name="firstName">
            {(field) => (
              <div className="space-y-2">
                <Label htmlFor={field.name}>First Name</Label>
                <Input
                  id={field.name}
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="John"
                />
              </div>
            )}
          </form.Field>

          <form.Field name="lastName">
            {(field) => (
              <div className="space-y-2">
                <Label htmlFor={field.name}>Last Name</Label>
                <Input
                  id={field.name}
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder="Doe"
                />
              </div>
            )}
          </form.Field>
        </div>

        <div className="flex gap-4 pt-4">
          <Button type="submit" className="flex-1">
            Save Changes
          </Button>
          <Button
            type="button"
            variant="outline"
            className="flex-1"
            onClick={() => router.history.back()}
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}

import { createFileRoute, redirect, useRouter } from '@tanstack/react-router';
import { useForm } from '@tanstack/react-form';
import { useId, useState } from 'react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

import {
  updateProfile as updateProfileApi,
  deleteProfile as deleteProfileApi,
} from '@/api/auth';
import { useAuthStore } from '@/store/auth';
import { generateDeviceId } from '@/lib/device';

export const Route = createFileRoute('/my-profile/edit')({
  component: EditAccountComponent,
  beforeLoad({ location }) {
    const { isSignedIn } = useAuthStore.getState();

    if (!isSignedIn) {
      throw redirect({
        to: '/auth/signin',
        search: { redirect: location.href },
      });
    }
  },
});

function EditAccountComponent() {
  const id = useId();
  const router = useRouter();

  const {
    activeProfile,
    updateProfile: updateProfileStore,
    deleteProfile: deleteProfileStore,
  } = useAuthStore();

  const profile = activeProfile();

  const [previewUrl, setPreviewUrl] = useState<string | null>(
    profile?.avatarUrl ?? null,
  );

  const form = useForm({
    defaultValues: {
      firstName: profile?.firstName ?? '',
      lastName: profile?.lastName ?? '',
      displayName: profile?.displayName ?? '',
      avatar: null as File | null,
    },
    onSubmit: async ({ value }) => {
      if (!profile?.id) return;

      const formData = new FormData();

      if (value.firstName && value.firstName !== profile.firstName) {
        formData.append('firstName', value.firstName);
      }

      if (value.lastName && value.lastName !== profile.lastName) {
        formData.append('lastName', value.lastName);
      }

      if (value.avatar) {
        formData.append('avatar', value.avatar);
      }

      formData.append(
        'metadata',
        JSON.stringify({
          deviceId: generateDeviceId(),
        }),
      );

      const res = await updateProfileApi(profile.id, formData);
      if (res.success) {
        updateProfileStore(profile.id, {
          firstName: value.firstName,
          lastName: value.lastName,
          avatarUrl: res.data.avatarUrl,
        });
        toast.success('Profile updated successfully');
        router.navigate({ to: '/my-account' });
      } else {
        toast.error(res.message);
      }
    },
  });

  const handleDeleteProfile = async () => {
    if (!profile?.id) return;

    const res = await deleteProfileApi(profile.id);
    if (!res.success) {
      toast.error(res.message);
      return;
    }

    deleteProfileStore(profile.id);
    router.navigate({ to: '/my-account' });
  };

  return (
    <div className="container mx-auto max-w-2xl py-10 px-4">
      <h1 className="mb-6 text-3xl font-bold">Edit Profile</h1>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
        className="space-y-8 rounded-lg border bg-card p-6 shadow-sm"
      >
        <div className="flex flex-col items-center gap-4 border-b pb-6">
          <Avatar className="h-24 w-24 border">
            <AvatarImage src={previewUrl ?? ''} />
            <AvatarFallback>
              {profile?.firstName?.[0]}
              {profile?.lastName?.[0]}
            </AvatarFallback>
          </Avatar>

          <form.Field name="avatar">
            {(field) => (
              <div className="flex flex-col items-center gap-2">
                <Label
                  htmlFor={id}
                  className="cursor-pointer rounded-md bg-secondary px-4 py-2 transition-colors hover:bg-secondary/80"
                >
                  Change Avatar
                </Label>
                <Input
                  id={id}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (!file) return;
                    field.handleChange(file);
                    setPreviewUrl(URL.createObjectURL(file));
                  }}
                />
              </div>
            )}
          </form.Field>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <form.Field name="firstName">
            {(field) => (
              <div className="space-y-2">
                <Label htmlFor={field.name}>First Name</Label>
                <Input
                  id={field.name}
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
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

        <Button
          type="button"
          variant="destructive"
          className="w-full"
          onClick={handleDeleteProfile}
        >
          Delete Profile
        </Button>
      </form>
    </div>
  );
}

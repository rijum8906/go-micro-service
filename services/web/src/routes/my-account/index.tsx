import { createFileRoute, Link, redirect } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useAuthStore } from '@/store/auth';

export const Route = createFileRoute('/my-account/')({
  component: MyAccountComponent,
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

function MyAccountComponent() {
  const { account, logout, activeProfile } = useAuthStore();
  const profile = activeProfile();

  if (!profile) return null;

  return (
    <div className="container mx-auto max-w-4xl py-10 px-4">
      <h1 className="mb-8 text-3xl font-bold">My Account</h1>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Profile Card */}
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-4">
              <Avatar className="h-16 w-16">
                <AvatarImage src={profile.avatarUrl ?? ''} />
                <AvatarFallback>
                  {profile.firstName[0]}
                  {profile.lastName?.[0]}
                </AvatarFallback>
              </Avatar>

              <div>
                <p className="text-lg font-semibold">
                  {profile.firstName} {profile.lastName}
                </p>
                <p className="text-sm text-muted-foreground">Active Profile</p>
              </div>
            </div>

            <div className="flex flex-wrap gap-3 pt-2">
              <Button asChild size="sm">
                <Link to="/my-profile/edit">Edit Profile</Link>
              </Button>

              {/* <Button asChild size="sm" variant="outline"> */}
              {/*   <Link to="/my-account/profiles">Switch Profile</Link> */}
              {/* </Button> */}
            </div>
          </CardContent>
        </Card>

        {/* Account Card */}
        <Card>
          <CardHeader>
            <CardTitle>Account</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm text-muted-foreground">Email</p>
              <p className="font-medium">{account?.email}</p>
            </div>

            <div className="flex flex-wrap gap-3 pt-2">
              {/* <Button asChild size="sm" variant="secondary"> */}
              {/*   <Link to="/my-account/settings">Account Settings</Link> */}
              {/* </Button> */}

              <Button asChild size="sm" variant="outline">
                <Link to="/my-account/edit/security">Security</Link>
              </Button>

              <Button size="sm" variant="destructive" onClick={logout}>
                Log out
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

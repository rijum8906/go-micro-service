import {
  ClientOnly,
  createFileRoute,
  Link,
  redirect,
} from '@tanstack/react-router';
import { Card, CardHeader } from '@/components/ui/card';
import { LoadingScreen } from '@/components/layout/loader';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/store/auth';

export const Route = createFileRoute('/dashboard/')({
  component: RouteComponent,
  beforeLoad(ctx) {
    const { isSignedIn } = useAuthStore.getState();

    if (!isSignedIn) {
      throw redirect({
        to: '/auth/signin',
        search: { redirect: ctx.location.href },
      });
    }
  },
});

function RouteComponent() {
  const { account, activeProfile } = useAuthStore();
  const profile = activeProfile();
  if (!profile) return null;

  return (
    <ClientOnly fallback={<LoadingScreen />}>
      <div className="p-6">
        <Card className="max-w-md">
          <CardHeader className="space-y-1">
            <div className="text-2xl font-bold">Dashboard</div>
            <div className="text-sm text-muted-foreground">
              Logged in as:{' '}
              <span className="font-medium">{account?.email}</span>
            </div>
          </CardHeader>

          <div className="p-6 pt-0">
            {profile ? (
              <div className="flex items-center gap-4">
                <div className="h-10 w-10 rounded-full bg-primary flex items-center justify-center text-white font-bold">
                  {profile.firstName}
                </div>
                <div>
                  <p className="text-sm font-medium leading-none">
                    {profile.firstName} {profile.lastName}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Active Profile
                  </p>
                </div>
                <div>
                  <Button>
                    <Link to="/my-account">My Account</Link>
                  </Button>
                </div>
              </div>
            ) : (
              <p className="text-sm text-yellow-600">No profile selected.</p>
            )}
          </div>
        </Card>
      </div>
    </ClientOnly>
  );
}

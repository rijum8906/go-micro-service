import {
  ClientOnly,
  createFileRoute,
  Link,
  redirect,
} from '@tanstack/react-router';
import { useAuthStore } from '@/store/auth';
import { Card, CardHeader } from '@/components/ui/card';
import { LoadingScreen } from '@/components/layout/loader';
import { Button } from '@/components/ui/button';

export const Route = createFileRoute('/dashboard/')({
  component: RouteComponent,
  beforeLoad(ctx) {
    const { isSignedIn, _hasHydrated } = useAuthStore.getState();

    if (_hasHydrated && !isSignedIn) {
      throw redirect({
        to: '/auth/signin',
        search: { redirect: ctx.location.href },
      });
    }
  },
});

function RouteComponent() {
  const { account, profiles, currentProfileIdx: currentIdx } = useAuthStore();

  const activeProfile =
    profiles && currentIdx !== null ? profiles[currentIdx] : null;
  console.log(activeProfile);

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
            {activeProfile ? (
              <div className="flex items-center gap-4">
                <div className="h-10 w-10 rounded-full bg-primary flex items-center justify-center text-white font-bold">
                  {activeProfile.first_name}
                </div>
                <div>
                  <p className="text-sm font-medium leading-none">
                    {activeProfile.first_name} {activeProfile.last_name}
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

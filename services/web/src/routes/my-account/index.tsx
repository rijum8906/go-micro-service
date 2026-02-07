import { createFileRoute, Link } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/store/auth';

export const Route = createFileRoute('/my-account/')({
  component: MyAccountComponent,
});

function MyAccountComponent() {
  const { account, profiles } = useAuthStore();

  return (
    <div className="container mx-auto py-10 px-4">
      <h1 className="text-3xl font-bold mb-6">My Account</h1>
      <div className="grid gap-6">
        <div className="p-6 border rounded-lg bg-card shadow-sm">
          <h2 className="text-xl font-semibold mb-4">Profile Information</h2>
          <div className="space-y-2">
            <p>
              <strong>Email:</strong> {account?.email}
            </p>
            <p>
              <strong>Name:</strong> {profiles?.[0]?.first_name}{' '}
              {profiles?.[0]?.last_name}
            </p>
          </div>
          <div className="mt-6">
            <Button asChild>
              <Link to="/my-account/edit">Edit Profile</Link>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

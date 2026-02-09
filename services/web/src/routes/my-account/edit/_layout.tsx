import {
  createFileRoute,
  Link,
  useLocation,
  Outlet,
} from '@tanstack/react-router';
import { User, Lock, Shield, Bell } from 'lucide-react';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';

export const Route = createFileRoute('/my-account/edit/_layout')({
  component: RouteComponent,
});

function RouteComponent() {
  const { pathname } = useLocation();

  return (
    <div className="container max-w-6xl py-10 px-4 mx-auto">
      {/* Header */}
      <div className="space-y-1">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground">
          Manage your account, security, and notification preferences.
        </p>
      </div>

      <Separator className="my-8" />

      <div className="grid grid-cols-1 lg:grid-cols-[240px_1fr] gap-10">
        {/* Sidebar */}
        <aside className="rounded-xl border bg-muted/30 p-2">
          <nav className="flex flex-col gap-1">
            <NavLink
              to="/my-account/edit/profile"
              icon={<User size={18} />}
              label="Profile"
              active={pathname.includes('profile')}
            />
            <NavLink
              to="/my-account/edit/security"
              icon={<Lock size={18} />}
              label="Security"
              active={pathname.includes('security')}
            />
            <NavLink
              to="/my-account/edit/privacy"
              icon={<Shield size={18} />}
              label="Privacy"
              active={pathname.includes('privacy')}
            />
            <NavLink
              to="/my-account/edit/notifications"
              icon={<Bell size={18} />}
              label="Notifications"
              active={pathname.includes('notifications')}
            />
          </nav>
        </aside>

        {/* Content */}
        <main className="rounded-xl border bg-background p-6 shadow-sm">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

function NavLink({
  to,
  icon,
  label,
  active,
}: {
  to: string;
  icon: React.ReactNode;
  label: string;
  active?: boolean;
}) {
  return (
    <Link to={to}>
      <div
        className={cn(
          'group relative flex items-center gap-3 rounded-lg px-4 py-2 text-sm font-medium transition-all',
          active
            ? 'bg-background shadow-sm text-foreground'
            : 'text-muted-foreground hover:bg-background hover:text-foreground',
        )}
      >
        {/* Active indicator */}
        <span
          className={cn(
            'absolute left-0 top-1/2 h-6 w-1 -translate-y-1/2 rounded-r-full transition-opacity',
            active ? 'bg-primary opacity-100' : 'opacity-0',
          )}
        />
        {icon}
        {label}
      </div>
    </Link>
  );
}

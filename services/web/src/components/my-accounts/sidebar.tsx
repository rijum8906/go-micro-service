import { Link, useLocation } from '@tanstack/react-router';
import { User, Lock, Shield, Bell } from 'lucide-react';
import { Separator } from '@/components/ui/separator';

export function Sidebar() {
  const { pathname } = useLocation();
  return (
    <div className="container max-w-6xl py-10 px-4 mx-auto">
      <div className="space-y-0.5">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground text-lg">
          Manage your account settings and set e-mail preferences.
        </p>
      </div>

      <Separator className="my-6" />

      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
        {/* Sidebar Navigation */}
        <aside className="lg:w-1/5">
          <nav className="flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1">
            <Link to="/my-account/edit/profile">
              <NavItem
                icon={<User size={18} />}
                label="Profile"
                active={pathname.includes('profile')}
              ></NavItem>
            </Link>
            <Link to="/my-account/edit/security">
              <NavItem
                icon={<Lock size={18} />}
                label="Security"
                active={pathname.includes('security')}
              />
            </Link>
            <Link to="/my-account/edit/privacy">
              <NavItem
                icon={<Shield size={18} />}
                label="Privacy"
                active={pathname.includes('privacy')}
              />
            </Link>
            <Link to="/my-account/edit/notifications">
              <NavItem
                icon={<Bell size={18} />}
                label="Notifications"
                active={pathname.includes('nofitications')}
              />
            </Link>
          </nav>
        </aside>
      </div>
    </div>
  );
}

function NavItem({
  icon,
  label,
  active = false,
}: {
  icon: React.ReactNode;
  label: string;
  active?: boolean;
}) {
  return (
    <button
      type="button"
      className={`flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
        active
          ? 'bg-primary text-primary-foreground shadow-sm'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      }`}
    >
      {icon}
      {label}
    </button>
  );
}

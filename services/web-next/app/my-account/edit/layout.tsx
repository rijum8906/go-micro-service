'use client'

import { usePathname } from 'next/navigation'
import Link from 'next/link'
import { User, Lock, Shield, Bell } from 'lucide-react'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'

function NavLink({
  href,
  icon,
  label,
  active,
}: {
  href: string
  icon: React.ReactNode
  label: string
  active?: boolean
}) {
  return (
    <Link href={href}>
      <div className={cn(
        'group relative flex items-center gap-3 rounded-lg px-4 py-2 text-sm font-medium transition-all',
        active
          ? 'bg-background shadow-sm text-foreground'
          : 'text-muted-foreground hover:bg-background hover:text-foreground',
      )}>
        <span className={cn(
          'absolute left-0 top-1/2 h-6 w-1 -translate-y-1/2 rounded-r-full transition-opacity',
          active ? 'bg-primary opacity-100' : 'opacity-0',
        )} />
        {icon}
        {label}
      </div>
    </Link>
  )
}

export default function EditLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  return (
    <div className="container max-w-6xl py-10 px-4 mx-auto">
      <div className="space-y-1">
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground">
          Manage your account, security, and notification preferences.
        </p>
      </div>
      <Separator className="my-8" />
      <div className="grid grid-cols-1 lg:grid-cols-[240px_1fr] gap-10">
        <aside className="rounded-xl border bg-muted/30 p-2">
          <nav className="flex flex-col gap-1">
            <NavLink href="/my-account/edit/profile" icon={<User size={18} />} label="Profile" active={pathname.includes('profile')} />
            <NavLink href="/my-account/edit/security" icon={<Lock size={18} />} label="Security" active={pathname.includes('security')} />
            <NavLink href="/my-account/edit/privacy" icon={<Shield size={18} />} label="Privacy" active={pathname.includes('privacy')} />
            <NavLink href="/my-account/edit/notifications" icon={<Bell size={18} />} label="Notifications" active={pathname.includes('notifications')} />
          </nav>
        </aside>
        <main className="rounded-xl border bg-background p-6 shadow-sm">
          {children}
        </main>
      </div>
    </div>
  )
}